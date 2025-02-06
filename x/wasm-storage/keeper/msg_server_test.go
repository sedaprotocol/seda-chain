package keeper_test

import (
	"encoding/hex"
	"os"

	"github.com/ethereum/go-ethereum/crypto"

	storetypes "cosmossdk.io/store/types"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

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
			name:   "happy path",
			preRun: func() {},
			input: types.MsgStoreOracleProgram{
				Sender: s.authority,
				Wasm:   regWasmZipped,
			},
			expErr: false,
			expOutput: types.MsgStoreOracleProgramResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(regWasm)),
			},
		},
		{
			name: "oracle program already exist",
			input: types.MsgStoreOracleProgram{
				Sender: s.authority,
				Wasm:   regWasmZipped,
			},
			preRun: func() {
				input := types.MsgStoreOracleProgram{
					Sender: s.authority,
					Wasm:   regWasmZipped,
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
				Sender: s.authority,
				Wasm:   regWasm,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "wasm is not gzip compressed",
		},
		{
			name: "oversized Wasm",
			input: types.MsgStoreOracleProgram{
				Sender: s.authority,
				Wasm:   oversizedWasmZipped,
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
			res, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
			if tc.expErr {
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expOutput, *res)
		})
	}
}

func (s *KeeperTestSuite) TestStoreOracleProgramGasMultiplier() {
	dummyWasm := make([]byte, 100)
	dummyWasmZipped, err := ioutils.GzipIt(dummyWasm)
	s.Require().NoError(err)

	s.SetupTest()

	// Start without a multiplier
	_, err = s.msgSrvr.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params: types.Params{
			MaxWasmSize:      1000000,
			WasmTTL:          1000,
			UploadMultiplier: 1,
		},
	})
	s.Require().NoError(err)

	// Use a fresh gas meter for the test
	firstMeter := storetypes.NewGasMeter(100000000000)
	_, err = s.msgSrvr.StoreOracleProgram(s.ctx.WithGasMeter(firstMeter), &types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   dummyWasmZipped,
	})
	s.Require().NoError(err)

	// We can't check the gas consumed here because for some reason it fluctuates between 12720 and 12750 :(

	// Set a multiplier of 200
	_, err = s.msgSrvr.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params: types.Params{
			MaxWasmSize:      1000000,
			WasmTTL:          1000,
			UploadMultiplier: 200,
		},
	})
	s.Require().NoError(err)

	// Change the dummy wasm slightly
	dummyWasm[0] = 0x01
	updatedDummyWasmZipped, err := ioutils.GzipIt(dummyWasm)
	s.Require().NoError(err)

	// Use a fresh gas meter for the test
	secondMeter := storetypes.NewGasMeter(100000000000)
	_, err = s.msgSrvr.StoreOracleProgram(s.ctx.WithGasMeter(secondMeter), &types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   updatedDummyWasmZipped,
	})
	s.Require().NoError(err)

	// We can't check the exact multiplier here because the gas consumed is not constant, and the multiplier
	// is only applied to the gas used to store the wasm, not the other operations in `StoreOracleProgram`.
	s.Require().Greater(secondMeter.GasConsumed(), firstMeter.GasConsumed()*uint64(100), "expected the gas consumed to be greater than 100x %d, got %d", firstMeter.GasConsumed(), secondMeter.GasConsumed())
}

func (s *KeeperTestSuite) TestStoreExecutorWasm() {
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
		input     types.MsgStoreExecutorWasm
		expErr    bool
		expErrMsg string
		expOutput types.MsgStoreExecutorWasmResponse
	}{
		{
			name: "happy path",
			input: types.MsgStoreExecutorWasm{
				Sender: s.authority,
				Wasm:   regWasmZipped,
			},
			preRun:    func() {},
			expErr:    false,
			expErrMsg: "",
			expOutput: types.MsgStoreExecutorWasmResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(regWasm)),
			},
		},
		{
			name: "invalid authority",
			input: types.MsgStoreExecutorWasm{
				Sender: "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Wasm:   regWasmZipped,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "executor wasm already exist",
			input: types.MsgStoreExecutorWasm{
				Sender: s.authority,
				Wasm:   regWasmZipped,
			},
			preRun: func() {
				input := types.MsgStoreExecutorWasm{
					Sender: s.authority,
					Wasm:   regWasmZipped,
				}
				_, err := s.msgSrvr.StoreExecutorWasm(s.ctx, &input)
				s.Require().NoError(err)
			},
			expErr:    true,
			expErrMsg: "wasm with the same hash already exists",
		},
		{
			name: "unzipped wasm",
			input: types.MsgStoreExecutorWasm{
				Sender: s.authority,
				Wasm:   regWasm,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "wasm is not gzip compressed",
		},
		{
			name: "oversized wasm",
			input: types.MsgStoreExecutorWasm{
				Sender: s.authority,
				Wasm:   oversizedWasmZipped,
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
			res, err := s.msgSrvr.StoreExecutorWasm(s.ctx, &input)
			if tc.expErr {
				s.Require().ErrorContains(err, tc.expErrMsg)
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
					MaxWasmSize:      1000000, // 1 MB
					WasmTTL:          100,
					UploadMultiplier: 10,
				},
			},
			expErrMsg: "",
		},
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Params: types.Params{
					MaxWasmSize:      1, // 1 MB
					WasmTTL:          1000,
					UploadMultiplier: 10,
				},
			},
			expErrMsg: "expected " + authority + ", got seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l: invalid authority",
		},
		{
			name: "invalid max wasm size",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize:      0, // 0 MB
					WasmTTL:          100,
					UploadMultiplier: 10,
				},
			},
			expErrMsg: "invalid max wasm size 0: invalid param",
		},
		{
			name: "invalid wasm time to live",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize:      111110,
					WasmTTL:          1,
					UploadMultiplier: 10,
				},
			},
			expErrMsg: "WasmTTL 1 < 2: invalid param",
		},
		{
			name: "invalid upload multiplier",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize:      111110,
					WasmTTL:          1000,
					UploadMultiplier: 0,
				},
			},
			expErrMsg: "invalid upload multiplier 0: invalid param",
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

// TODO(#347) Expiration is disabled for now.
/*
func (s *KeeperTestSuite) TestDRWasmPruning() {
	params, err := s.keeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	wasmTTL := params.WasmTTL

	// Get the list of all oracle programs.
	oraclePrograms := s.keeper.ListOraclePrograms(s.ctx)
	s.Require().Empty(oraclePrograms)

	// Save 1 DR Wasm with default exp [params.WasmTTL]
	drWasm1, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	drWasmZipped1, err := ioutils.GzipIt(drWasm1)
	s.Require().NoError(err)

	resp1, err := s.msgSrvr.StoreOracleProgram(s.ctx, &types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   drWasmZipped1,
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

	resp2, err := s.msgSrvr.StoreOracleProgram(s.ctx, &types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   drWasmZipped2,
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
	list := s.keeper.ListOraclePrograms(s.ctx)

	s.Require().ElementsMatch(list, []string{fmt.Sprintf("%s,%d", resp1.Hash, firstWasmPruneHeight), fmt.Sprintf("%s,%d", resp2.Hash, secondWasmPruneHeight)})
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
	list = s.keeper.ListOraclePrograms(s.ctx)
	s.Require().ElementsMatch(list, []string{fmt.Sprintf("%s,%d", resp2.Hash, secondWasmPruneHeight)})
	// Check WsmExp is in sync
	s.Require().Len(getAllWasmExpEntry(s.T(), s.ctx, s.keeper), 1)

	// H = 2 * params.WasmTTL.
	s.ctx = s.ctx.WithBlockHeight(secondWasmPruneHeight)
	list = s.keeper.ListOraclePrograms(s.ctx)
	s.Require().ElementsMatch(list, []string{fmt.Sprintf("%s,%d", resp2.Hash, secondWasmPruneHeight)})
	// Simulate EndBlocker Call
	s.Require().NoError(s.keeper.EndBlock(s.ctx))

	// Go to the next block
	s.ctx = s.ctx.WithBlockHeight(secondWasmPruneHeight + 1)

	// Both wasm must be pruned.
	list = s.keeper.ListOraclePrograms(s.ctx)
	s.Require().Empty(list) // Check WsmExp is in sync
	s.Require().Empty(getAllWasmExpEntry(s.T(), s.ctx, s.keeper))
}

func getAllWasmExpEntry(t *testing.T, c sdk.Context, k *keeper.Keeper) []string {
	t.Helper()
	it, err := k.OracleProgramExpiration.Iterate(c, nil)
	require.NoError(t, err)
	keys, err := it.Keys()
	require.NoError(t, err)
	hashes := make([]string, 0)
	for _, key := range keys {
		hashes = append(hashes, hex.EncodeToString(key.K2()))
	}
	return hashes
}
*/
