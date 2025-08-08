package keeper_test

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/exp/rand"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/require"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var (
	testAddrs = []sdk.AccAddress{
		sdk.AccAddress([]byte("to0_________________")),
		sdk.AccAddress([]byte("to1_________________")),
		sdk.AccAddress([]byte("to2_________________")),
	}
)

func (s *KeeperTestSuite) TestStoreOracleProgram() {
	regWasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	regWasmZipped, err := ioutils.GzipIt(regWasm)
	s.Require().NoError(err)

	oversizedWasm, err := os.ReadFile("testutil/oversized.wasm")
	s.Require().NoError(err)
	oversizedWasmZipped, err := ioutils.GzipIt(oversizedWasm)
	s.Require().NoError(err)

	cases := []struct {
		name      string
		preRun    func()
		input     types.MsgStoreOracleProgram
		expErr    bool
		expErrMsg string
		expOutput types.MsgStoreOracleProgramResponse
	}{
		{
			name: "happy path",
			preRun: func() {
				s.ApplyDefaultMockExpectations()
			}, input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       regWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			expErr: false,
			expOutput: types.MsgStoreOracleProgramResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(regWasm)),
			},
		},
		{
			name: "Invalid address",
			preRun: func() {
				s.ApplyDefaultMockExpectations()
			}, input: types.MsgStoreOracleProgram{
				Sender:     "seda9ucvn84ynyjzyzeavwvurmdyxat26l",
				Wasm:       regWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			expErr:    true,
			expErrMsg: "invalid address",
		},
		{
			name: "Invalid storage fee",
			preRun: func() {
				s.ApplyDefaultMockExpectations()
			}, input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       regWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(math.MaxInt64).Add(sdkmath.NewInt(1)).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			expErr:    true,
			expErrMsg: "WASM file is too large",
		},
		{
			name: "Storage fee is too low",
			preRun: func() {
				s.ApplyDefaultMockExpectations()
			},
			input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       regWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm)-100)).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			expErr:    true,
			expErrMsg: "exceeds limit",
		},
		{
			name: "Insufficient funds",
			preRun: func() {
				s.mockBankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), authtypes.FeeCollectorName, gomock.Any()).Return(errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, "insufficient funds")).Times(1)
				s.ApplyDefaultMockExpectations()
			},
			input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       regWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
		{
			name: "oracle program already exist",
			input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       regWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			preRun: func() {
				s.ApplyDefaultMockExpectations()
				input := types.MsgStoreOracleProgram{
					Sender:     s.authority,
					Wasm:       regWasmZipped,
					StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
				}
				_, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
				s.Require().Nil(err)
			},
			expErr:    true,
			expErrMsg: "already exists",
		},
		{
			name: "unzipped Wasm",
			input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       regWasm,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(regWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			preRun: func() {
				s.ApplyDefaultMockExpectations()
			},
			expErr:    true,
			expErrMsg: "wasm is not gzip compressed",
		},
		{
			name: "oversized Wasm",
			input: types.MsgStoreOracleProgram{
				Sender:     s.authority,
				Wasm:       oversizedWasmZipped,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(int64(len(oversizedWasm))).Mul(sdkmath.NewInt(int64(types.DefaultWasmCostPerByte))))),
			},
			preRun: func() {
				s.ApplyDefaultMockExpectations()
			},
			expErr:    true,
			expErrMsg: "",
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			input := tc.input
			res, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
			if tc.expErr {
				s.Require().Contains(err.Error(), tc.expErrMsg)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expOutput, *res)
		})
	}
}

func (s *KeeperTestSuite) TestInstantiateCoreContract() {
	cases := []struct {
		name      string
		input     *types.MsgInstantiateCoreContract
		expErrMsg string
	}{
		{
			name: "happy path",
			input: &types.MsgInstantiateCoreContract{
				Sender: s.authority,
				Admin:  s.authority,
				CodeID: 1,
				Label:  "label",
				Msg:    []byte(`{}`),
				Funds:  sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(1000000000000000000))),
				Salt:   []byte("salt"),
				FixMsg: true,
			},
			expErrMsg: "",
		},
		{
			name: "invalid authority",
			input: &types.MsgInstantiateCoreContract{
				Sender: "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Admin:  s.authority,
				CodeID: 1,
				Label:  "label",
				Msg:    []byte(`{}`),
				Funds:  sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(1000000000000000000))),
				Salt:   []byte("salt"),
				FixMsg: true,
			},
			expErrMsg: "expected " + s.authority + ", got seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l: invalid authority",
		},
		{
			name: "invalid msg json",
			input: &types.MsgInstantiateCoreContract{
				Sender: s.authority,
				Admin:  s.authority,
				CodeID: 1,
				Label:  "label",
				Msg:    []byte(`}`),
				Funds:  sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewInt(1000000000000000000))),
				Salt:   []byte("salt"),
				FixMsg: true,
			},
			expErrMsg: "invalid",
		},
	}

	s.SetupTest()
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.mockWasmKeeper.EXPECT().Instantiate2(
				s.ctx, tc.input.CodeID, gomock.Any(), gomock.Any(),
				tc.input.Msg, tc.input.Label, tc.input.Funds, tc.input.Salt,
				tc.input.FixMsg).Return(testAddrs[0], []byte{}, nil).MaxTimes(1)

			_, err := s.msgSrvr.InstantiateCoreContract(s.ctx, tc.input)
			if tc.expErrMsg != "" {
				s.Require().Error(err)
				s.Require().Equal(tc.expErrMsg, err.Error())
				return
			}
			s.Require().NoError(err)

			// Check that the address is correctly set.
			addr, err := s.keeper.CoreContractRegistry.Get(s.ctx)
			s.Require().NoError(err)
			s.Require().Equal(testAddrs[0].String(), addr)
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParams() {
	authority := s.keeper.GetAuthority()
	cases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErrMsg string
	}{
		{
			name: "happy path",
			input: &types.MsgUpdateParams{
				Authority: s.authority,
				Params: types.Params{
					MaxWasmSize:     1000000, // 1 MB
					WasmCostPerByte: 50000000000000,
				},
			},
			expErrMsg: "",
		},
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Params: types.Params{
					MaxWasmSize:     1, // 1 MB
					WasmCostPerByte: 50000000000000,
				},
			},
			expErrMsg: "expected " + authority + ", got seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l: invalid authority",
		},
		{
			name: "invalid max wasm size",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize:     0, // 0 MB
					WasmCostPerByte: 50000000000000,
				},
			},
			expErrMsg: "invalid max wasm size 0: invalid param",
		},
		{
			name: "invalid cost per byte",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize:     111110,
					WasmCostPerByte: 0,
				},
			},
			expErrMsg: "invalid wasm cost per byte 0: invalid param",
		},
	}

	s.SetupTest()
	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := s.msgSrvr.UpdateParams(s.ctx, tc.input)
			if tc.expErrMsg != "" {
				s.Require().Error(err)
				s.Require().Equal(tc.expErrMsg, err.Error())
				return
			}
			s.Require().NoError(err)

			// Check that the Params were correctly set
			params, _ := s.keeper.Params.Get(s.ctx)
			s.Require().Equal(tc.input.Params, params)
		})
	}
}

func (s *KeeperTestSuite) TestRefundTxFee() {
	s.SetupTest()

	coreContract := "seda1p9dtxynydns46wgv7wthsgh4rhvp7cw8rn99a6ekga7jzufqkhcqgrgcwg"
	err := s.keeper.CoreContractRegistry.Set(s.ctx, coreContract)
	s.Require().NoError(err)

	cases := []struct {
		name            string
		isLastMsgReveal bool
		input           *types.MsgRefundTxFee
		gasLimit        uint64
		refund          bool
		expErrMsg       string
	}{
		{
			name: "not the core contract",
			input: &types.MsgRefundTxFee{
				Authority: "seda13uj26mq00terh0lrrhd3vajt723c03xsp6nu3stmpf7t5l36vlsq7zpct2",
				DrId:      "drID",
				PublicKey: "pubKey",
				IsReveal:  false,
			},
			gasLimit:  2000,
			refund:    false,
			expErrMsg: "expected " + coreContract + ", got seda13uj26mq00terh0lrrhd3vajt723c03xsp6nu3stmpf7t5l36vlsq7zpct2: invalid authority",
		},
		{
			name: "not the last message's DR ID",
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drIDa",
				PublicKey: "pubKey",
				IsReveal:  false,
			},
			gasLimit: 2000,
			refund:   false,
		},
		{
			name: "not the last message's public key",
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drID",
				PublicKey: "pubKeya",
				IsReveal:  false,
			},
			gasLimit: 2000,
			refund:   false,
		},
		{
			name:            "last msg is reveal, not commit",
			isLastMsgReveal: true,
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drID",
				PublicKey: "pubKey",
				IsReveal:  false,
			},
			gasLimit: 2000,
			refund:   false,
		},
		{
			name:            "last msg is commit, not reveal",
			isLastMsgReveal: false,
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drID",
				PublicKey: "pubKey",
				IsReveal:  true,
			},
			gasLimit: 2000,
			refund:   false,
		},
		{
			name: "gas limit too large for refund",
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drID",
				PublicKey: "pubKey",
				IsReveal:  false,
			},
			gasLimit: 10000,
			refund:   false,
		},
		{
			name:            "happy path - last msg is commit",
			isLastMsgReveal: false,
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drID",
				PublicKey: "pubKey",
				IsReveal:  false,
			},
			gasLimit: 2000,
			refund:   true,
		},
		{
			name:            "happy path - last msg is reveal",
			isLastMsgReveal: true,
			input: &types.MsgRefundTxFee{
				Authority: coreContract,
				DrId:      "drID",
				PublicKey: "pubKey",
				IsReveal:  true,
			},
			gasLimit: 2000,
			refund:   true,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			fee := sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewIntFromUint64(tc.gasLimit*1e10)))
			txBytes := generateTxBytes(s.T(), s.txConfig, coreContract, fee, rand.Intn(10)+1, tc.isLastMsgReveal)

			s.ctx = s.ctx.WithTxBytes(txBytes)
			s.ctx = s.ctx.WithGasMeter(storetypes.NewGasMeter(tc.gasLimit))

			if tc.refund {
				s.mockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), authtypes.FeeCollectorName, testAddrs[0], fee).Return(nil).Times(1)
			}

			_, err = s.msgSrvr.RefundTxFee(s.ctx, tc.input)
			if tc.expErrMsg != "" {
				s.Require().Error(err)
				s.Require().Equal(tc.expErrMsg, err.Error())
				return
			}
			s.Require().NoError(err)
		})
	}
}

// generateTxBytes generates numMsg commit/reveal messages, with the last message
// having fixed data request ID and public key ("drID" and "pubKey").
func generateTxBytes(t *testing.T, txConfig client.TxConfig, coreContract string, fee sdk.Coins, numMsgs int, isLastMsgReveal bool) []byte {
	var msgs []sdk.Msg
	for i := 0; i < numMsgs; i++ {
		var msg []byte
		if i == numMsgs-1 {
			if isLastMsgReveal {
				msg = testutil.RevealMsg("drID", "reveal", "pubKey", "proof", []string{}, 0, 99, 777)
			} else {
				msg = testutil.CommitMsg("drID", "commitment", "pubKey", "proof")
			}
		} else {
			if i%2 == 0 {
				msg = testutil.CommitMsg(fmt.Sprintf("drID%d", i), "commitment", fmt.Sprintf("pubKey%d", i), "proof")
			} else {
				msg = testutil.RevealMsg(fmt.Sprintf("drID%d", i), "reveal", fmt.Sprintf("pubKey%d", i), "proof", []string{}, 0, 99, 777)
			}
		}
		contractMsg := wasmtypes.MsgExecuteContract{
			Sender:   sdk.AccAddress(testAddrs[0]).String(),
			Contract: coreContract,
			Msg:      msg,
			Funds:    sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, sdkmath.NewIntFromUint64(1))),
		}
		msgs = append(msgs, &contractMsg)
	}

	txf := tx.Factory{}.
		WithChainID("chain-id").
		WithTxConfig(txConfig).
		WithFees(fee.String()).
		WithFeePayer(sdk.AccAddress(testAddrs[0]))
	tx, err := txf.BuildUnsignedTx(msgs...)
	require.NoError(t, err)

	txBytes, err := txConfig.TxEncoder()(tx.GetTx())
	require.NoError(t, err)
	return txBytes
}
