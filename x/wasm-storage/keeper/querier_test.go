package keeper_test

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestDataRequestWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	input := types.MsgStoreDataRequestWasm{
		Sender:   s.authority,
		Wasm:     compWasm,
		WasmType: types.WasmTypeDataRequest,
	}
	storedWasm, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
	s.Require().NoError(err)

	req := types.QueryDataRequestWasmRequest{Hash: storedWasm.Hash}
	res, err := s.queryClient.DataRequestWasm(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, hex.EncodeToString(res.Wasm.Hash))
}

func (s *KeeperTestSuite) TestOverlayWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	input := types.MsgStoreOverlayWasm{
		Sender:   s.authority,
		Wasm:     compWasm,
		WasmType: types.WasmTypeDataRequestExecutor,
	}
	storedWasm, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input)
	s.Require().NoError(err)

	req := types.QueryOverlayWasmRequest{Hash: storedWasm.Hash}
	res, err := s.queryClient.OverlayWasm(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, hex.EncodeToString(res.Wasm.Hash))
}

func (s *KeeperTestSuite) TestDataRequestWasms() {
	s.SetupTest()
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	input := types.MsgStoreDataRequestWasm{
		Sender:   s.authority,
		Wasm:     compWasm,
		WasmType: types.WasmTypeDataRequest,
	}
	storedWasm, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
	s.Require().NoError(err)

	wasm2, err := os.ReadFile("test_utils/cowsay.wasm")
	s.Require().NoError(err)
	compWasm2, err := ioutils.GzipIt(wasm2)
	s.Require().NoError(err)
	input2 := types.MsgStoreDataRequestWasm{
		Sender:   s.authority,
		Wasm:     compWasm2,
		WasmType: types.WasmTypeDataRequest,
	}
	storedWasm2, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input2)
	s.Require().NoError(err)

	req := types.QueryDataRequestWasmsRequest{}
	res, err := s.queryClient.DataRequestWasms(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(fmt.Sprintf("%s,%s", storedWasm.Hash, "WASM_TYPE_DATA_REQUEST"), res.HashTypePairs[0])
	s.Require().Equal(fmt.Sprintf("%s,%s", storedWasm2.Hash, "WASM_TYPE_DATA_REQUEST"), res.HashTypePairs[1])
}

func (s *KeeperTestSuite) TestOverlayWasms() {
	s.SetupTest()
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	input := types.MsgStoreOverlayWasm{
		Sender:   s.authority,
		Wasm:     compWasm,
		WasmType: types.WasmTypeRelayer,
	}
	storedWasm, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input)
	s.Require().NoError(err)

	wasm2, err := os.ReadFile("test_utils/cowsay.wasm")
	s.Require().NoError(err)
	compWasm2, err := ioutils.GzipIt(wasm2)
	s.Require().NoError(err)
	input2 := types.MsgStoreOverlayWasm{
		Sender:   s.authority,
		Wasm:     compWasm2,
		WasmType: types.WasmTypeRelayer,
	}
	storedWasm2, err := s.msgSrvr.StoreOverlayWasm(s.ctx, &input2)
	s.Require().NoError(err)

	req := types.QueryOverlayWasmsRequest{}
	res, err := s.queryClient.OverlayWasms(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(fmt.Sprintf("%s,%s", storedWasm.Hash, "WASM_TYPE_RELAYER"), res.HashTypePairs[0])
	s.Require().Equal(fmt.Sprintf("%s,%s", storedWasm2.Hash, "WASM_TYPE_RELAYER"), res.HashTypePairs[1])
}
