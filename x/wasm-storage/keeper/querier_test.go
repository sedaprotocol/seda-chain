package keeper_test

import (
	"encoding/hex"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestOracleProgram() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	input := types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   compWasm,
	}
	storedWasm, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
	s.Require().NoError(err)

	req := types.QueryOracleProgramRequest{Hash: storedWasm.Hash}
	res, err := s.queryClient.OracleProgram(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(storedWasm.Hash, hex.EncodeToString(res.OracleProgram.Hash))
}

func (s *KeeperTestSuite) TestOraclePrograms() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	input := types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   compWasm,
	}
	storedWasm, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input)
	s.Require().NoError(err)

	wasm2, err := os.ReadFile("testutil/cowsay.wasm")
	s.Require().NoError(err)
	compWasm2, err := ioutils.GzipIt(wasm2)
	s.Require().NoError(err)
	input2 := types.MsgStoreOracleProgram{
		Sender: s.authority,
		Wasm:   compWasm2,
	}
	storedWasm2, err := s.msgSrvr.StoreOracleProgram(s.ctx, &input2)
	s.Require().NoError(err)

	req := types.QueryOracleProgramsRequest{}
	res, err := s.queryClient.OraclePrograms(s.ctx, &req)
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Contains(res.List[0], storedWasm.Hash)
	s.Require().Contains(res.List[1], storedWasm2.Hash)
}
