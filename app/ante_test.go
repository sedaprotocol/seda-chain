package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/math"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkcdctestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/sedaprotocol/seda-chain/app/params"
	appparams "github.com/sedaprotocol/seda-chain/app/params"
	apptestutil "github.com/sedaprotocol/seda-chain/app/testutil"
	"github.com/sedaprotocol/seda-chain/testutil"
	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	wasmstoragekeeper "github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var testAddrs = []sdk.AccAddress{
	sdk.AccAddress([]byte("to0_________________")),
	sdk.AccAddress([]byte("to1_________________")),
	sdk.AccAddress([]byte("to2_________________")),
}

// mockAnteHandler implements a simple AnteHandler for testing
func mockAnteHandler(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	return ctx, nil
}

func TestCommitRevealDecorator_AnteHandle(t *testing.T) {
	ctx := sdk.Context{}
	interfaceRegistry := sdkcdctestutil.CodecOptions{
		AccAddressPrefix: params.Bech32PrefixAccAddr,
		ValAddressPrefix: params.Bech32PrefixValAddr,
	}.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(codec, tx.DefaultSignModes)

	coreContractAddr := sdk.AccAddress(testAddrs[0])
	bondDenom := appparams.DefaultBondDenom
	sender := testAddrs[0]

	ctrl := gomock.NewController(t)
	wasmStorageKeeper := apptestutil.NewMockWasmStorageKeeper(ctrl)
	decorator := NewCommitRevealDecorator(wasmStorageKeeper)

	wasmStorageKeeper.EXPECT().GetCoreContractAddr(gomock.Any()).Return(coreContractAddr, nil).AnyTimes()

	tests := []struct {
		name          string
		setupMock     func(*wasmstoragekeeper.Keeper)
		msgs          []sdk.Msg
		expectedError string
	}{
		{
			name: "happy path - mix of contract and module msgs",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id_2",
					PublicKey: "public_key",
					Proof:     "proof",
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.RevealMsg("dr_id_2", "reveal", "public_key", "proof", []string{}, 0, 99, 35000),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgReveal{
					Sender: sender.String(),
					RevealBody: &coretypes.RevealBody{
						DrID: "dr_id",
					},
					PublicKey: "public_key",
					Proof:     "proof",
				},
			},
			expectedError: "",
		},
		{
			name: "happy path - only contract msgs",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.RevealMsg("dr_id", "reveal", "public_key", "proof", []string{}, 0, 99, 35000),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
			},
			expectedError: "",
		},
		{
			name: "happy path - only module msgs",
			msgs: []sdk.Msg{
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id",
					PublicKey: "public_key",
					Proof:     "proof",
				},
				&coretypes.MsgReveal{
					Sender: sender.String(),
					RevealBody: &coretypes.RevealBody{
						DrID: "dr_id",
					},
					PublicKey: "public_key",
					Proof:     "proof",
				},
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id",
					PublicKey: "public_key_2",
					Proof:     "proof",
				},
				&coretypes.MsgReveal{
					Sender: sender.String(),
					RevealBody: &coretypes.RevealBody{
						DrID: "dr_id",
					},
					PublicKey: "public_key_2",
					Proof:     "proof",
				},
			},
			expectedError: "",
		},
		{
			name: "duplicate commits",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id",
					PublicKey: "public_key",
					Proof:     "proof",
				},
			},
			expectedError: "duplicate commit or reveal message detected",
		},
		{
			name: "duplicate reveals",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.RevealMsg("dr_id", "reveal", "public_key", "proofproof", []string{"b"}, 1, 5, 27000),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.RevealMsg("dr_id_2", "reveal", "public_key", "proof", []string{"a"}, 0, 99, 35000),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgReveal{
					Sender: sender.String(),
					RevealBody: &coretypes.RevealBody{
						DrID: "dr_id",
					},
					PublicKey: "public_key",
					Proof:     "proof",
				},
			},
			expectedError: "duplicate commit or reveal message detected",
		},
		{
			name: "non-commit/reveal msg appears in the front",
			msgs: []sdk.Msg{
				&wasmstoragetypes.MsgStoreOracleProgram{
					Sender:     testAddrs[1].String(),
					Wasm:       []byte{},
					StorageFee: sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id_2",
					PublicKey: "public_key",
					Proof:     "proof",
				},
			},
			expectedError: "commit or reveal message cannot be mixed with other messages",
		},
		{
			name: "non-commit/reveal msg appears in the middle",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&wasmstoragetypes.MsgStoreOracleProgram{
					Sender:     testAddrs[1].String(),
					Wasm:       []byte{},
					StorageFee: sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id_2",
					PublicKey: "public_key",
					Proof:     "proof",
				},
			},
			expectedError: "commit or reveal message cannot be mixed with other messages",
		},
		{
			name: "non-commit/reveal msg appears at the end",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.RevealMsg("dr_id", "reveal", "public_key", "proof", []string{}, 0, 99, 35000),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgReveal{
					Sender: sender.String(),
					RevealBody: &coretypes.RevealBody{
						DrID: "dr_id_2",
					},
					PublicKey: "public_key",
					Proof:     "proof",
				},
				&wasmstoragetypes.MsgStoreOracleProgram{
					Sender:     testAddrs[1].String(),
					Wasm:       []byte{},
					StorageFee: sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
			},
			expectedError: "commit or reveal message cannot be mixed with other messages",
		},
		{
			name: "only non-commit/reveal messages",
			msgs: []sdk.Msg{
				&wasmstoragetypes.MsgStoreOracleProgram{
					Sender:     testAddrs[1].String(),
					Wasm:       []byte{},
					StorageFee: sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&pubkeytypes.MsgAddKey{
					ValidatorAddr:  "seda1qyqszqgpqyqszqgpqyqszqgpqyqszqgp",
					IndexedPubKeys: []pubkeytypes.IndexedPubKey{},
				},
			},
			expectedError: "",
		},
		{
			name: "mix with nil commit message",
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      []byte(`{"commit_data_result": null}`),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
				&coretypes.MsgCommit{
					Sender:    sender.String(),
					DrID:      "dr_id",
					PublicKey: "public_key_2",
					Proof:     "proof",
				},
				&wasmtypes.MsgExecuteContract{
					Sender:   sender.String(),
					Contract: coreContractAddr.String(),
					Msg:      testutil.CommitMsg("dr_id", "commitment", "public_key", "proof"),
					Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
				},
			},
			expectedError: "commit or reveal message cannot be mixed with other messages",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fee := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(20000*1e10)))
			txf := clienttx.Factory{}.
				WithChainID("ante-test").
				WithTxConfig(txConfig).
				WithFees(fee.String()).
				WithFeePayer(sender)
			tx, err := txf.BuildUnsignedTx(tc.msgs...)
			require.NoError(t, err)

			txBytes, err := txConfig.TxEncoder()(tx.GetTx())
			require.NoError(t, err)
			ctx = ctx.WithTxBytes(txBytes)

			_, err = decorator.AnteHandle(ctx, tx.GetTx(), false, mockAnteHandler)

			if tc.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}
