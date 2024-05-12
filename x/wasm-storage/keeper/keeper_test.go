package keeper_test

import (
	"encoding/hex"
	"math/rand"
	"os"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestSetDataRequestWasm() {
	s.SetupTest()
	wasm, err := os.ReadFile("test_utils/hello-world.wasm")
	s.Require().NoError(err)
	compWasm, err := ioutils.GzipIt(wasm)
	s.Require().NoError(err)
	mockWasm := wasmstoragetypes.Wasm{
		Hash:     crypto.Keccak256(compWasm),
		Bytecode: compWasm,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	s.Require().NoError(s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm), mockWasm))
}

func (s *KeeperTestSuite) TestGetDataRequestWasm() {
	s.SetupTest()
	mockWasm := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	err := s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm), mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.DataRequestWasm.Get(s.ctx, keeper.WasmKey(mockWasm))
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasDataRequestWasm() {
	s.SetupTest()
	mockWasm := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	has, _ := s.keeper.DataRequestWasm.Has(s.ctx, keeper.WasmKey(mockWasm))
	s.Assert().False(has)
	err := s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm), mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.DataRequestWasm.Has(s.ctx, keeper.WasmKey(mockWasm))
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestSetOverlayWasm() {
	s.SetupTest()
	mockWasm := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	s.Require().NoError(s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasm), mockWasm))
}

func (s *KeeperTestSuite) TestGetOverlayWasm() {
	s.SetupTest()
	mockWasm := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	err := s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasm), mockWasm)
	s.Require().NoError(err)
	value, _ := s.keeper.OverlayWasm.Get(s.ctx, keeper.WasmKey(mockWasm))
	s.Assert().NotNil(value)
	s.Assert().Equal(mockWasm, value)
}

func (s *KeeperTestSuite) TestHasOverlayWasm() {
	s.SetupTest()
	mockWasm := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	has, _ := s.keeper.OverlayWasm.Has(s.ctx, keeper.WasmKey(mockWasm))
	s.Assert().False(has)
	err := s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasm), mockWasm)
	s.Require().NoError(err)
	has, _ = s.keeper.OverlayWasm.Has(s.ctx, keeper.WasmKey(mockWasm))
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestIterateAllDataRequestWasm() {
	s.SetupTest()
	mockWasm1 := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	mockWasm2 := wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}

	err := s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm1), mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm2), mockWasm2)
	s.Require().NoError(err)
	err = s.keeper.IterateAllDataRequestWasms(s.ctx, func(wasm wasmstoragetypes.Wasm) (stop bool) {
		s.Assert().Equal(wasmstoragetypes.WasmTypeDataRequest, wasm.GetWasmType())
		return true
	})
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestListDateRequestWasm() {
	s.SetupTest()
	mockWasm1 := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	mockWasm2 := wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}

	err := s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm1), mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm2), mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListDataRequestWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Equal(hex.EncodeToString(mockWasm1.Hash)+","+mockWasm1.WasmType.String(), result[0])
	s.Assert().Equal(hex.EncodeToString(mockWasm2.Hash)+","+mockWasm2.WasmType.String(), result[1])
}

func (s *KeeperTestSuite) TestListOverlayWasm() {
	s.SetupTest()
	mockWasm1 := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	mockWasm2 := wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}

	err := s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasm1), mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasm2), mockWasm2)
	s.Require().NoError(err)
	result := s.keeper.ListOverlayWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Equal(hex.EncodeToString(mockWasm1.Hash)+","+mockWasm1.WasmType.String(), result[0])
	s.Assert().Equal(hex.EncodeToString(mockWasm2.Hash)+","+mockWasm2.WasmType.String(), result[1])
}

func (s *KeeperTestSuite) TestGetAllWasms() {
	s.SetupTest()
	mockWasm1 := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	mockWasm2 := wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}

	err := s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm1), mockWasm1)
	s.Require().NoError(err)
	err = s.keeper.DataRequestWasm.Set(s.ctx, keeper.WasmKey(mockWasm2), mockWasm2)
	s.Require().NoError(err)

	mockWasmO1 := wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	mockWasmO2 := wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}

	err = s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasmO1), mockWasmO1)
	s.Require().NoError(err)
	err = s.keeper.OverlayWasm.Set(s.ctx, keeper.WasmKey(mockWasmO2), mockWasmO2)
	s.Require().NoError(err)

	result := s.keeper.GetAllWasms(s.ctx)
	s.Assert().Equal(4, len(result))
	s.Assert().Equal(mockWasm1, result[0])
	s.Assert().Equal(mockWasm2, result[1])
	s.Assert().Equal(mockWasmO1, result[2])
	s.Assert().Equal(mockWasmO2, result[3])
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
		mockWasm := wasmstoragetypes.NewWasm(byteCode, wasmstoragetypes.WasmTypeDataRequest, tm, bh, ttl)
		wasmKey := keeper.WasmKey(*mockWasm)
		err := s.keeper.DataRequestWasm.Set(s.ctx, wasmKey, *mockWasm)
		s.Require().NoError(err)
		expKey := collections.Join(expHeight, wasmKey)
		err = s.keeper.WasmExp.Set(s.ctx, expKey)
		s.Require().NoError(err)

		wasmKeys = append(wasmKeys, wasmKey)
	}

	result, err := s.keeper.WasmKeyByExpBlock(s.ctx, expHeight)
	s.Require().NoError(err)
	s.Require().ElementsMatch(wasmKeys, result)
}
