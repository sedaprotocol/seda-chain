package keeper_test

import (
	"encoding/hex"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestDataRequestWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	input := types.MsgStoreDataRequestWasm{
		Sender: s.authority,
		Wasm:   compWasm,
	}
	storedWasm, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
	s.Require().NoError(err)

	req := types.QueryDataRequestWasmRequest{Hash: storedWasm.Hash}
	res, err := s.queryClient.DataRequestWasm(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, hex.EncodeToString(res.Wasm.Hash))
}

func (s *KeeperTestSuite) TestExecutorWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	input := types.MsgStoreExecutorWasm{
		Sender: s.authority,
		Wasm:   compWasm,
	}
	storedWasm, err := s.msgSrvr.StoreExecutorWasm(s.ctx, &input)
	s.Require().NoError(err)

	req := types.QueryExecutorWasmRequest{Hash: storedWasm.Hash}
	res, err := s.queryClient.ExecutorWasm(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, hex.EncodeToString(res.Wasm.Hash))
}

func (s *KeeperTestSuite) TestDataRequestWasms() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	input := types.MsgStoreDataRequestWasm{
		Sender: s.authority,
		Wasm:   compWasm,
	}
	storedWasm, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input)
	s.Require().NoError(err)

	wasm2, err := os.ReadFile("testutil/cowsay.wasm")
	s.Require().NoError(err)
	compWasm2, err := ioutils.GzipIt(wasm2)
	s.Require().NoError(err)
	input2 := types.MsgStoreDataRequestWasm{
		Sender: s.authority,
		Wasm:   compWasm2,
	}
	storedWasm2, err := s.msgSrvr.StoreDataRequestWasm(s.ctx, &input2)
	s.Require().NoError(err)

	req := types.QueryDataRequestWasmsRequest{}
	res, err := s.queryClient.DataRequestWasms(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Contains(res.List[0], storedWasm.Hash)
	s.Require().Contains(res.List[1], storedWasm2.Hash)
}

func (s *KeeperTestSuite) TestExecutorWasms() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	input := types.MsgStoreExecutorWasm{
		Sender: s.authority,
		Wasm:   compWasm,
	}
	storedWasm, err := s.msgSrvr.StoreExecutorWasm(s.ctx, &input)
	s.Require().NoError(err)

	wasm2, err := os.ReadFile("testutil/cowsay.wasm")
	s.Require().NoError(err)
	compWasm2, err := ioutils.GzipIt(wasm2)
	s.Require().NoError(err)
	input2 := types.MsgStoreExecutorWasm{
		Sender: s.authority,
		Wasm:   compWasm2,
	}
	storedWasm2, err := s.msgSrvr.StoreExecutorWasm(s.ctx, &input2)
	s.Require().NoError(err)

	req := types.QueryExecutorWasmsRequest{}
	res, err := s.queryClient.ExecutorWasms(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, res.List[0])
	s.Require().Equal(storedWasm2.Hash, res.List[1])
}
