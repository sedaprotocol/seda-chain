package keeper_test

import (
	"encoding/hex"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestSetOracleProgram() {
	s.SetupTest()
	compWasm, err := ioutils.GzipIt(testwasms.HelloWorldWasm())
	s.Require().NoError(err)

	mockWasm := types.NewOracleProgram(compWasm, time.Now().UTC())
	s.Require().NoError(s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm))
}

func (s *KeeperTestSuite) TestGetOracleProgram() {
	s.SetupTest()
	mockWasm := types.NewOracleProgram(mockedByteArray, time.Now().UTC())
	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.OracleProgram.Get(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasOracleProgram() {
	s.SetupTest()
	mockWasm := types.NewOracleProgram(mockedByteArray, time.Now().UTC())
	has, _ := s.keeper.OracleProgram.Has(s.ctx, mockWasm.Hash)
	s.Assert().False(has)
	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.OracleProgram.Has(s.ctx, mockWasm.Hash)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestIterateOracleProgram() {
	s.SetupTest()
	mockWasm1 := types.NewOracleProgram(mockedByteArray, time.Now().UTC())
	mockWasm2 := types.NewOracleProgram(append(mockedByteArray, 2), time.Now().UTC())
	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.OracleProgram.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)

	var results []types.OracleProgram
	err = s.keeper.IterateOraclePrograms(s.ctx, func(wasm types.OracleProgram) (stop bool) {
		results = append(results, wasm)
		return false
	})
	s.Assert().ElementsMatch([]types.OracleProgram{mockWasm1, mockWasm2}, results)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestListDateRequestWasm() {
	s.SetupTest()
	mockWasm1 := types.NewOracleProgram(mockedByteArray, time.Now().UTC())
	mockWasm2 := types.NewOracleProgram(append(mockedByteArray, 2), time.Now().UTC())

	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.OracleProgram.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListOraclePrograms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm1.Hash))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm2.Hash))
}
