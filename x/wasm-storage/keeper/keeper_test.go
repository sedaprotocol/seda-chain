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

func (s *KeeperTestSuite) TestSetDataRequestWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("testutil/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)

	mockWasm := types.NewDataRequestWasm(compWasm, time.Now().UTC(), 1000, 100)
	s.Require().NoError(s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm))
}

func (s *KeeperTestSuite) TestGetDataRequestWasm() {
	s.SetupTest()
	mockWasm := types.NewDataRequestWasm(mockedByteArray, time.Now().UTC(), 1000, 100)
	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.DataRequestWasm.Get(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasDataRequestWasm() {
	s.SetupTest()
	mockWasm := types.NewDataRequestWasm(mockedByteArray, time.Now().UTC(), 1000, 100)
	has, _ := s.keeper.DataRequestWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().False(has)
	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.DataRequestWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestSetExecutorWasm() {
	s.SetupTest()
	mockWasm := types.NewExecutorWasm(mockedByteArray, time.Now().UTC())
	s.Require().NoError(s.keeper.ExecutorWasm.Set(s.ctx, mockWasm.Hash, mockWasm))
}

func (s *KeeperTestSuite) TestGetExecutorWasm() {
	s.SetupTest()
	mockWasm := types.NewExecutorWasm(mockedByteArray, time.Now().UTC())
	err := s.keeper.ExecutorWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.ExecutorWasm.Get(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasExecutorWasm() {
	s.SetupTest()
	mockWasm := types.NewExecutorWasm(mockedByteArray, time.Now().UTC())
	has, _ := s.keeper.ExecutorWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().False(has)
	err := s.keeper.ExecutorWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.ExecutorWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestIterateDataRequestWasm() {
	s.SetupTest()
	mockWasm1 := types.NewDataRequestWasm(mockedByteArray, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewDataRequestWasm(append(mockedByteArray, 2), time.Now().UTC(), 1000, 100)
	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)

	var results []types.DataRequestWasm
	err = s.keeper.IterateDataRequestWasms(s.ctx, func(wasm types.DataRequestWasm) (stop bool) {
		results = append(results, wasm)
		return false
	})
	s.Assert().ElementsMatch([]types.DataRequestWasm{mockWasm1, mockWasm2}, results)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestIterateExecutorWasm() {
	s.SetupTest()
	mockWasm1 := types.NewExecutorWasm(mockedByteArray, time.Now().UTC())
	mockWasm2 := types.NewExecutorWasm(append(mockedByteArray, 2), time.Now().UTC())
	err := s.keeper.ExecutorWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.ExecutorWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)

	var results []types.ExecutorWasm
	err = s.keeper.IterateExecutorWasms(s.ctx, func(wasm types.ExecutorWasm) (stop bool) {
		results = append(results, wasm)
		return false
	})
	s.Assert().ElementsMatch([]types.ExecutorWasm{mockWasm1, mockWasm2}, results)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestListDateRequestWasm() {
	s.SetupTest()
	mockWasm1 := types.NewDataRequestWasm(mockedByteArray, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewDataRequestWasm(append(mockedByteArray, 2), time.Now().UTC(), 1000, 100)

	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListDataRequestWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm1.Hash)+","+strconv.FormatInt(mockWasm1.ExpirationHeight, 10))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm2.Hash)+","+strconv.FormatInt(mockWasm2.ExpirationHeight, 10))
}

func (s *KeeperTestSuite) TestListExecutorWasm() {
	s.SetupTest()
	mockWasm1 := types.NewExecutorWasm(mockedByteArray, time.Now().UTC())
	mockWasm2 := types.NewExecutorWasm(append(mockedByteArray, 2), time.Now().UTC())

	err := s.keeper.ExecutorWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.ExecutorWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListExecutorWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm1.Hash))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm2.Hash))
}

func (s *KeeperTestSuite) TestGetAllWasms() {
	s.SetupTest()
	mockWasm1 := types.NewDataRequestWasm(mockedByteArray, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewDataRequestWasm(append(mockedByteArray, 2), time.Now().UTC(), 1000, 100)

	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)

	mockWasm3 := types.NewExecutorWasm(mockedByteArray, time.Now().UTC())
	mockWasm4 := types.NewExecutorWasm(append(mockedByteArray, 2), time.Now().UTC())

	err = s.keeper.ExecutorWasm.Set(s.ctx, mockWasm3.Hash, mockWasm3)
	s.Require().NoError(err)
	err = s.keeper.ExecutorWasm.Set(s.ctx, mockWasm4.Hash, mockWasm4)
	s.Require().NoError(err)

	res1 := s.keeper.GetAllDataRequestWasms(s.ctx)
	s.Assert().Equal(2, len(res1))
	s.Assert().Contains(res1, mockWasm1)
	s.Assert().Contains(res1, mockWasm2)

	res2 := s.keeper.GetAllExecutorWasms(s.ctx)
	s.Assert().Equal(2, len(res2))
	s.Assert().Contains(res2, mockWasm3)
	s.Assert().Contains(res2, mockWasm4)
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
		mockWasm := types.NewDataRequestWasm(byteCode, tm, bh, ttl)

		err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
		s.Require().NoError(err)

		expKey := collections.Join(expHeight, mockWasm.Hash)
		err = s.keeper.WasmExpiration.Set(s.ctx, expKey)
		s.Require().NoError(err)

		wasmKeys = append(wasmKeys, mockWasm.Hash)
	}

	result, err := s.keeper.GetExpiredWasmKeys(s.ctx, expHeight)
	s.Require().NoError(err)
	s.Require().ElementsMatch(wasmKeys, result)
}
