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
					UploadMultiplier: 10,
				},
			},
			expErrMsg: "invalid max wasm size 0: invalid param",
		},
		{
			name: "invalid upload multiplier",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MaxWasmSize:      111110,
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
