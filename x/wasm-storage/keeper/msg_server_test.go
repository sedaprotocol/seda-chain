package keeper_test

import (
	"encoding/hex"
	"math"
	"os"

	"go.uber.org/mock/gomock"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
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
