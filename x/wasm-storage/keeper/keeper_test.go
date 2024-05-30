package keeper_test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"cosmossdk.io/collections"
	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestSetDataRequestWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	mockWasm := types.Wasm{
		Hash:     crypto.Keccak256(compWasm),
		Bytecode: compWasm,
		WasmType: types.WasmTypeDataRequest,
	}
	s.Require().NoError(s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm))
}

func (s *KeeperTestSuite) TestGetDataRequestWasm() {
	s.SetupTest()
	mockWasm := types.NewWasm(mockedByteArray, types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)
	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.DataRequestWasm.Get(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasDataRequestWasm() {
	s.SetupTest()
	mockWasm := types.NewWasm(mockedByteArray, types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)
	has, _ := s.keeper.DataRequestWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().False(has)
	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.DataRequestWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestSetOverlayWasm() {
	s.SetupTest()
	mockWasm := types.NewWasm(mockedByteArray, types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)
	s.Require().NoError(s.keeper.OverlayWasm.Set(s.ctx, mockWasm.Hash, mockWasm))
}

func (s *KeeperTestSuite) TestGetOverlayWasm() {
	s.SetupTest()
	mockWasm := types.NewWasm(mockedByteArray, types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)
	err := s.keeper.OverlayWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.OverlayWasm.Get(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasOverlayWasm() {
	s.SetupTest()
	mockWasm := types.NewWasm(mockedByteArray, types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)
	has, _ := s.keeper.OverlayWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().False(has)
	err := s.keeper.OverlayWasm.Set(s.ctx, mockWasm.Hash, mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.OverlayWasm.Has(s.ctx, mockWasm.Hash)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestIterateAllDataRequestWasm() {
	s.SetupTest()
	mockWasm1 := types.NewWasm(mockedByteArray, types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewWasm(append(mockedByteArray, 2), types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)

	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	err = s.keeper.IterateAllDataRequestWasms(s.ctx, func(wasm types.Wasm) (stop bool) {
		s.Assert().Equal(types.WasmTypeDataRequest, wasm.GetWasmType())
		return true
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestListDateRequestWasm() {
	s.SetupTest()
	mockWasm1 := types.NewWasm(mockedByteArray, types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewWasm(append(mockedByteArray, 2), types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)

	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListDataRequestWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm1.Hash)+","+mockWasm1.WasmType.String())
	s.Assert().Contains(result, hex.EncodeToString(mockWasm2.Hash)+","+mockWasm2.WasmType.String())
}

func (s *KeeperTestSuite) TestListOverlayWasm() {
	s.SetupTest()
	mockWasm1 := types.NewWasm(mockedByteArray, types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewWasm(append(mockedByteArray, 2), types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)

	err := s.keeper.OverlayWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.OverlayWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListOverlayWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Contains(result, hex.EncodeToString(mockWasm1.Hash)+","+mockWasm1.WasmType.String())
	s.Assert().Contains(result, hex.EncodeToString(mockWasm2.Hash)+","+mockWasm2.WasmType.String())
}

func (s *KeeperTestSuite) TestGetAllWasms() {
	s.SetupTest()
	mockWasm1 := types.NewWasm(mockedByteArray, types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)
	mockWasm2 := types.NewWasm(append(mockedByteArray, 2), types.WasmTypeDataRequest, time.Now().UTC(), 1000, 100)

	err := s.keeper.DataRequestWasm.Set(s.ctx, mockWasm1.Hash, mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, mockWasm2.Hash, mockWasm2)
	s.Require().NoError(err)

	mockWasm3 := types.NewWasm(mockedByteArray, types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)
	mockWasm4 := types.NewWasm(append(mockedByteArray, 2), types.WasmTypeRelayer, time.Now().UTC(), 1000, 100)

	err = s.keeper.OverlayWasm.Set(s.ctx, mockWasm3.Hash, mockWasm3)
	s.Require().NoError(err)
	err = s.keeper.OverlayWasm.Set(s.ctx, mockWasm4.Hash, mockWasm4)
	s.Require().NoError(err)

	result := s.keeper.GetAllWasms(s.ctx)
	s.Assert().Equal(4, len(result))
	s.Assert().Contains(result, mockWasm1)
	s.Assert().Contains(result, mockWasm2)
	s.Assert().Contains(result, mockWasm3)
	s.Assert().Contains(result, mockWasm4)
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
		mockWasm := types.NewWasm(byteCode, types.WasmTypeDataRequest, tm, bh, ttl)

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

func (s *KeeperTestSuite) TestKeeper_GasConsumeToStoreDR() {
	s.SetupTest()
	size := map[string]int64{
		"1 KB":   1000,
		"500 KB": 500000,
		"1 MB":   1000000,
		"1 GB":   1e+9,
	}
	for name, ln := range size {
		wasmData := make([]byte, ln)
		file, err := os.Create(name + ".bin")
		s.Require().NoError(err)

		fmt.Printf("%s: %d\n", name, binary.Size(wasmData))
		wasm := wasmstoragetypes.NewWasm(wasmData, wasmstoragetypes.WasmTypeDataRequest, s.ctx.BlockTime(), s.ctx.BlockHeight(), 10000)
		s.Require().NoError(s.keeper.DataRequestWasm.Set(s.ctx, wasm.Hash, wasm))

		//bgm := s.ctx.BlockGasMeter()
		//fmt.Printf("Gas: %v\n", bgm.GasConsumed())
		s.Require().NoError(binary.Write(file, binary.BigEndian, wasmData))
		s.Require().NoError(file.Close())
	}
}
