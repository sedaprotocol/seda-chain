package keeper_test

import (
	"encoding/hex"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/hyperledger/burrow/crypto"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestStoreDataRequestWasm() {
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)

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
				Wasm:     compWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			expErr: false,
			expOutput: types.MsgStoreDataRequestWasmResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(wasm)),
			},
		},
		{
			name: "Overlay wasm already exist",
			input: types.MsgStoreDataRequestWasm{
				Sender:   s.authority,
				Wasm:     compWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun: func() {
				input := types.MsgStoreDataRequestWasm{
					Sender:   s.authority,
					Wasm:     compWasm,
					WasmType: types.WasmTypeDataRequest,
				}
				_, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
				s.Require().Nil(err)
			},
			expErr:    true,
			expErrMsg: "Data Request Wasm with given hash already exists",
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			res, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &tc.input)
			if tc.expErr {
				s.Require().Error(err, tc.expErrMsg)
			} else {
				s.Require().Nil(err)
				s.Require().Equal(tc.expOutput, *res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestStoreOverlayWasm() {
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
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
				Wasm:     compWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun:    func() {},
			expErr:    false,
			expErrMsg: "",
			expOutput: types.MsgStoreOverlayWasmResponse{
				Hash: hex.EncodeToString(crypto.Keccak256(wasm)),
			},
		},
		{
			name: "invalid authority",
			input: types.MsgStoreOverlayWasm{
				Sender:   "this-is-not-valid",
				Wasm:     compWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun:    func() {},
			expErr:    true,
			expErrMsg: "invalid authority this-is-not-valid",
		},
		{
			name: "Overlay wasm already exist",
			input: types.MsgStoreOverlayWasm{
				Sender:   s.authority,
				Wasm:     compWasm,
				WasmType: types.WasmTypeDataRequest,
			},
			preRun: func() {
				input := types.MsgStoreOverlayWasm{
					Sender:   s.authority,
					Wasm:     compWasm,
					WasmType: types.WasmTypeDataRequest,
				}
				_, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input)
				s.Require().Nil(err)
			},
			expErr:    true,
			expErrMsg: "Overlay Wasm with given hash already exists",
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.preRun()
			res, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &tc.input)
			if tc.expErr {
				s.Require().Error(err, tc.expErrMsg)
			} else {
				s.Require().Nil(err)
				s.Require().Equal(tc.expOutput, *res)
			}
		})
	}
}
