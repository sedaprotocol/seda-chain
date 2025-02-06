package keeper_test

import (
	"encoding/hex"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestSetOracleProgram() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	mockWasm := types.NewOracleProgram(compWasm, time.Now().UTC(), 1000, 100)
	s.Require().NoError(s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm))
}

func (s *KeeperTestSuite) TestGetOracleProgram() {
	s.SetupTest()
	mockWasm := types.NewOracleProgram(mockedByteArray, time.Now().UTC(), 1000, 100)
	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.OracleProgram.Get(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasOracleProgram() {
	s.SetupTest()
	mockWasm := types.NewOracleProgram(mockedByteArray, time.Now().UTC(), 1000, 100)
	has, _ := s.keeper.OracleProgram.Has(s.ctx, mockWasm.Hash)
	s.Assert().False(has)
	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.OracleProgram.Has(s.ctx, mockWasm.Hash)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestIterateOracleProgram() {
	s.SetupTest()
	mockWasm1 := types.NewOracleProgram(mockedByteArray, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewOracleProgram(append(mockedByteArray, 2), time.Now().UTC(), 1000, 100)
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
	mockWasm1 := types.NewOracleProgram(mockedByteArray, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewOracleProgram(append(mockedByteArray, 2), time.Now().UTC(), 1000, 100)

	err := s.keeper.OracleProgram.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.OracleProgram.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListOraclePrograms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm1.Hash)+","+strconv.FormatInt(mockWasm1.ExpirationHeight, 10))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm2.Hash)+","+strconv.FormatInt(mockWasm2.ExpirationHeight, 10))
}

func (s *KeeperTestSuite) TestKeeper_WasmKeyByExpBlock() {
	s.SetupTest()
	N := rand.Intn(100) + 1
	ttl := int64(rand.Intn(100000000))
	tm := s.ctx.BlockTime()
	bh := s.ctx.BlockHeight()
	expHeight := bh + ttl
	wasmKeys := make([][]byte, 0, N)
	for i := 0; i < N; i++ {
		byteCode := append(mockedByteArray, byte(i)) //nolint: gocritic
		mockWasm := types.NewOracleProgram(byteCode, tm, bh, ttl)

		err := s.keeper.OracleProgram.Set(s.ctx, mockWasm.Hash, mockWasm)
		s.Require().NoError(err)

		expKey := collections.Join(expHeight, mockWasm.Hash)
		err = s.keeper.OracleProgramExpiration.Set(s.ctx, expKey)
		s.Require().NoError(err)

		wasmKeys = append(wasmKeys, mockWasm.Hash)
	}

	result, err := s.keeper.GetExpiredOracleProgamKeys(s.ctx, expHeight)
	s.Require().NoError(err)
	s.Require().ElementsMatch(wasmKeys, result)
}
