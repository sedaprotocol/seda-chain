package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestStoreDataRequestWasm() {
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
		input     types.MsgStoreDataRequestWasm
		expErr    bool
		expErrMsg string
		expOutput types.MsgStoreDataRequestWasmResponse
	}{
		{
			name:   "happy path",
			preRun: func() {},
			input: types.MsgStoreDataRequestWasm{
				Sender:   s.authority,
				Wasm:     regWasmZipped,
				WasmType: types.WasmTypeDataRequest,
			},
			expErr: false,
			expOutput: types.MsgStoreDataRequestWasmResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(regWasm)),
			},
		},
		{
			name: "Data Request wasm already exist",
			input: types.MsgStoreDataRequestWasm{
				Sender:   s.authority,
				Wasm:     regWasmZipped,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun: func() {
				input := types.MsgStoreDataRequestWasm{
					Sender:   s.authority,
					Wasm:     regWasmZipped,
					WasmType: types.WasmTypeDataRequest,
				}
				_, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
				s.Require().Nil(err)
			},
			expErr:    true,
			expErrMsg: "already exists",
		},
		{
			name: "unzipped Wasm",
			input: types.MsgStoreDataRequestWasm{
				Sender:   s.authority,
				Wasm:     regWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "wasm is not gzip compressed",
		},
		{
			name: "oversized Wasm",
			input: types.MsgStoreDataRequestWasm{
				Sender:   s.authority,
				Wasm:     oversizedWasmZipped,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "",
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			input := tc.input
			res, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
			if tc.expErr {
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expOutput, *res)
		})
	}
}

func (s *KeeperTestSuite) TestStoreOverlayWasm() {
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
		input     types.MsgStoreOverlayWasm
		expErr    bool
		expErrMsg string
		expOutput types.MsgStoreOverlayWasmResponse
	}{
		{
			name: "happy path",
			input: types.MsgStoreOverlayWasm{
				Sender:   s.authority,
				Wasm:     regWasmZipped,
				WasmType: types.WasmTypeRelayer,
			},
			preRun:    func() {},
			expErr:    false,
			expErrMsg: "",
			expOutput: types.MsgStoreOverlayWasmResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(regWasm)),
			},
		},
		{
			name: "invalid wasm type",
			input: types.MsgStoreOverlayWasm{
				Sender:   s.authority,
				Wasm:     regWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "overlay Wasm type must be data-request-executor or relayer",
		},
		{
			name: "invalid authority",
			input: types.MsgStoreOverlayWasm{
				Sender:   "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Wasm:     regWasmZipped,
				WasmType: types.WasmTypeRelayer,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "Overlay wasm already exist",
			input: types.MsgStoreOverlayWasm{
				Sender:   s.authority,
				Wasm:     regWasmZipped,
				WasmType: types.WasmTypeRelayer,
			},
			preRun: func() {
				input := types.MsgStoreOverlayWasm{
					Sender:   s.authority,
					Wasm:     regWasmZipped,
					WasmType: types.WasmTypeRelayer,
				}
				_, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input)
				s.Require().NoError(err)
			},
			expErr:    true,
			expErrMsg: "overlay Wasm with given hash already exists",
		},
		{
			name: "unzipped Wasm",
			input: types.MsgStoreOverlayWasm{
				Sender:   s.authority,
				Wasm:     regWasm,
				WasmType: types.WasmTypeRelayer,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "wasm is not gzip compressed",
		},
		{
			name: "oversized Wasm",
			input: types.MsgStoreOverlayWasm{
				Sender:   s.authority,
				Wasm:     oversizedWasmZipped,
				WasmType: types.WasmTypeRelayer,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "",
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			input := tc.input
			res, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input)
			if tc.expErr {
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expOutput, *res)
		})
	}
}

func (s *KeeperTestSuite) TestMarshalJSON() {
	cases := []struct {
		name     string
		hash     string
		body     *types.EventStoreDataRequestWasm
		expected string
	}{
		{
			name: "Test WasmTypeDataRequest",
			hash: "8558424e10c60eb4594cb2f1de834d5dd7a3b073d98d8641f8985fdbd84c3261",
			body: &types.EventStoreDataRequestWasm{
				Hash:     "8558424e10c60eb4594cb2f1de834d5dd7a3b073d98d8641f8985fdbd84c3261",
				WasmType: types.WasmTypeDataRequest,
				Bytecode: []byte("test WasmTypeDataRequest"),
			},
			expected: `{"hash":"8558424e10c60eb4594cb2f1de834d5dd7a3b073d98d8641f8985fdbd84c3261","wasm_type":1,"bytecode":"dGVzdCBXYXNtVHlwZURhdGFSZXF1ZXN0"}`,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			wrapper := &keeper.EventStoreDataRequestWasmWrapper{
				EventStoreDataRequestWasm: &types.EventStoreDataRequestWasm{
					Hash:     tc.hash,
					WasmType: tc.body.WasmType,
					Bytecode: tc.body.Bytecode,
				},
			}

			data, err := json.Marshal(wrapper)
			s.Require().NoError(err)
			s.Require().Equal(tc.expected, string(data))
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParams() {
	authority := s.keeper.GetAuthority()
	cases := []struct {
		name      string
		input     types.MsgUpdateParams
		expErrMsg string
	}{
		{
			name: "happy path",
			input: types.MsgUpdateParams{
				Authority: s.authority,
				Params: types.Params{
					MaxWasmSize: 1000000, // 1 MB
					WasmTTL:     100,
				},
			},
			expErrMsg: "",
		},
		{
			name: "invalid authority",
			input: types.MsgUpdateParams{
				Authority: "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Params: types.Params{
					MaxWasmSize: 1, // 1 MB
					WasmTTL:     1000,
				},
			},
			expErrMsg: "invalid authority; expected " + authority + ", got seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
		},
		{
			name: "invalid max wasm size",
			input: types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize: 0, // 0 MB
					WasmTTL:     100,
				},
			},
			expErrMsg: "invalid max Wasm size: 0",
		},
		{
			name: "invalid wasm time to live",
			input: types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize: 111110,
					WasmTTL:     1,
				},
			},
			expErrMsg: "WasmTTL 1 < 2: invalid param",
		},
	}

	s.SetupTest()
	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := s.msgSrvr.UpdateParams(s.ctx, &tc.input)
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

func (s *KeeperTestSuite) TestDRWasmPruning() {
	params, err := s.keeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	wasmTTL := params.WasmTTL

	// Get the list of all Data Request Wasms
	dataRequestWasms := s.keeper.ListDataRequestWasms(s.ctx)
	s.Require().Empty(dataRequestWasms)

	// Save 1 DR Wasm with default exp [params.WasmTTL]
	drWasm1, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	drWasmZipped1, err := ioutils.GzipIt(drWasm1)
	s.Require().NoError(err)

	resp1, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &types.MsgStoreDataRequestWasm{
		Sender:   s.authority,
		Wasm:     drWasmZipped1,
		WasmType: types.WasmTypeDataRequest,
	})
	s.Require().NoError(err)

	// Save 1 DR Wasm with default 2 * exp [params.WasmTTL]
	// First double the wasm lifespan.
	params.WasmTTL = 2 * wasmTTL
	s.Require().NoError(s.keeper.Params.Set(s.ctx, params))

	drWasm2, err := os.ReadFile("testutil/cowsay.wasm")
	s.Require().NoError(err)
	drWasmZipped2, err := ioutils.GzipIt(drWasm2)
	s.Require().NoError(err)

	resp2, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &types.MsgStoreDataRequestWasm{
		Sender:   s.authority,
		Wasm:     drWasmZipped2,
		WasmType: types.WasmTypeDataRequest,
	})
	s.Require().NoError(err)

	firstWasmPruneHeight := s.ctx.BlockHeight() + wasmTTL
	secondWasmPruneHeight := s.ctx.BlockHeight() + (2 * wasmTTL)

	// Wasm pruning takes place during the EndBlocker. If the height of a pruning block is H,
	// and the wasm to prune is W;
	// W would be available at H. W would NOT be available from H+1.

	// Artificially move to the pruning block for the first wasm.
	s.ctx = s.ctx.WithBlockHeight(firstWasmPruneHeight)

	// H = params.WasmTTL. || firstWasmPruneHeight => 0 + params.WasmTTL
	// We still have 2 wasms.
	drWasmHash := s.keeper.ListDataRequestWasms(s.ctx)
	s.Require().ElementsMatch(drWasmHash, []string{resp1.Hash + ",WASM_TYPE_DATA_REQUEST", resp2.Hash + ",WASM_TYPE_DATA_REQUEST"})
	// Check WsmExp is in sync
	s.Require().Len(getAllWasmExpEntry(s.T(), s.ctx, s.keeper), 2)

	// Simulate EndBlocker Call. This will remove one wasm.
	s.Require().NoError(s.keeper.EndBlock(s.ctx))

	// Go to the next block
	// H = params.WasmTTL + 1.
	s.ctx = s.ctx.WithBlockHeight(firstWasmPruneHeight + 1)
	// Simulate EndBlocker Call. This EndBlocker call will have no effect. As at this height no wasm to prune.
	s.Require().NoError(s.keeper.EndBlock(s.ctx))
	// Check: 1 wasm was pruned, 1 remained.
	drWasmHash = s.keeper.ListDataRequestWasms(s.ctx)
	s.Require().ElementsMatch(drWasmHash, []string{resp2.Hash + ",WASM_TYPE_DATA_REQUEST"})
	// Check WsmExp is in sync
	s.Require().Len(getAllWasmExpEntry(s.T(), s.ctx, s.keeper), 1)

	// H = 2 * params.WasmTTL.
	s.ctx = s.ctx.WithBlockHeight(secondWasmPruneHeight)
	drWasmHash = s.keeper.ListDataRequestWasms(s.ctx)
	s.Require().ElementsMatch(drWasmHash, []string{resp2.Hash + ",WASM_TYPE_DATA_REQUEST"})
	// Simulate EndBlocker Call
	s.Require().NoError(s.keeper.EndBlock(s.ctx))

	// Go to the next block
	s.ctx = s.ctx.WithBlockHeight(secondWasmPruneHeight + 1)

	// Both wasm must be pruned.
	drWasmHash = s.keeper.ListDataRequestWasms(s.ctx)
	s.Require().Empty(drWasmHash) // Check WsmExp is in sync
	s.Require().Empty(getAllWasmExpEntry(s.T(), s.ctx, s.keeper))
}

func getAllWasmExpEntry(t *testing.T, c sdk.Context, k *keeper.Keeper) []string {
	t.Helper()
	it, err := k.WasmExpiration.Iterate(c, nil)
	require.NoError(t, err)
	keys, err := it.Keys()
	require.NoError(t, err)
	hashes := make([]string, 0)
	for _, key := range keys {
		hashes = append(hashes, hex.EncodeToString(key.K2()))
	}
	return hashes
}
