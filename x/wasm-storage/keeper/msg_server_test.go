package keeper_test

import (
	"encoding/hex"
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestStoreDataRequestWasm() {
	regWasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	regWasmZipped, err := ioutils.GzipIt(regWasm)
	s.Require().NoError(err)

	oversizedWasm, err := os.ReadFile("test_utils/oversized.wasm")
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
			expErrMsg: "data Request Wasm with given hash already exists",
		},
		// TO-DO: Add after migrating ValidateBasic logic
		// {
		// 	name: "inconsistent Wasm type",
		// 	input: types.MsgStoreDataRequestWasm{
		// 		Sender:   s.authority,
		// 		Wasm:     regWasmZipped,
		// 		WasmType: types.WasmTypeRelayer,
		// 	},
		// 	preRun:    func() {},
		// 	expErr:    true,
		// 	expErrMsg: "not a Data Request Wasm",
		// },
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
	for i := range cases {
		tc := cases[i]
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			input := tc.input
			res, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
			if tc.expErr {
				s.Require().ErrorContains(err, tc.expErrMsg)
			} else {
				s.Require().Nil(err)
				s.Require().Equal(tc.expOutput, *res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestStoreOverlayWasm() {
	regWasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	regWasmZipped, err := ioutils.GzipIt(regWasm)
	s.Require().NoError(err)

	oversizedWasm, err := os.ReadFile("test_utils/oversized.wasm")
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
				Sender:   "cosmos16wfryel63g7axeamw68630wglalcnk3l0zuadc",
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
				s.Require().Nil(err)
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
	for i := range cases {
		tc := cases[i]
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			input := tc.input
			res, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input)
			if tc.expErr {
				s.Require().ErrorContains(err, tc.expErrMsg)
			} else {
				s.Require().Nil(err)
				s.Require().Equal(tc.expOutput, *res)
			}
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

func (s *KeeperTestSuite) TestSetParams() {
	cases := []struct {
		name      string
		preRun    func()
		input     types.Params
		expErr    bool
		expErrMsg string
	}{
		{
			name: "happy path",
			input: types.Params{
				MaxWasmSize: 1000000, // 1 MB
			},
			preRun:    func() {},
			expErr:    false,
			expErrMsg: "",
		},
		// negative cases would be caught in ValidateBasic before reaching keeper
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			err := s.wasmStorageKeeper.SetParams(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Equal(tc.expErrMsg, err.Error())
			} else {
				s.Require().NoError(err)

				// Check that the Params were correctly set
				params := s.wasmStorageKeeper.GetParams(s.ctx)
				s.Require().Equal(tc.input, params)
			}
		})
	}
}
