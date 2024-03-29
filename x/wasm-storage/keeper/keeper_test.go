package keeper_test

import (
	"encoding/hex"
	"os"

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
	mockWasm := &wasmstoragetypes.Wasm{
		Hash:     crypto.Keccak256(compWasm),
		Bytecode: compWasm,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm)
}

func (s *KeeperTestSuite) TestGetDataRequestWasm() {
	s.SetupTest()
	mockWasm := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm)
	value := s.wasmStorageKeeper.GetDataRequestWasm(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(*mockWasm, *value)
}

func (s *KeeperTestSuite) TestHasDataRequestWasm() {
	s.SetupTest()
	mockWasm := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	has := s.wasmStorageKeeper.HasDataRequestWasm(s.ctx, mockWasm)
	s.Assert().False(has)
	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm)
	has = s.wasmStorageKeeper.HasDataRequestWasm(s.ctx, mockWasm)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestSetOverlayWasm() {
	s.SetupTest()
	mockWasm := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasm)
}

func (s *KeeperTestSuite) TestGetOverlayWasm() {
	s.SetupTest()
	mockWasm := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasm)
	value := s.wasmStorageKeeper.GetOverlayWasm(s.ctx, mockWasm.Hash)
	s.Assert().NotNil(value)
	s.Assert().Equal(*mockWasm, *value)
}

func (s *KeeperTestSuite) TestHasOverlayWasm() {
	s.SetupTest()
	mockWasm := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	has := s.wasmStorageKeeper.HasOverlayWasm(s.ctx, mockWasm)
	s.Assert().False(has)
	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasm)
	has = s.wasmStorageKeeper.HasOverlayWasm(s.ctx, mockWasm)
	s.Assert().True(has)
}

func (s *KeeperTestSuite) TestIterateAllDataRequestWasm() {
	s.SetupTest()
	mockWasm1 := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	mockWasm2 := &wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}

	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm1)
	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm2)
	s.wasmStorageKeeper.IterateAllDataRequestWasms(s.ctx, func(wasm wasmstoragetypes.Wasm) (stop bool) {
		s.Assert().Equal(wasmstoragetypes.WasmTypeDataRequest, wasm.GetWasmType())
		return true
	})
}

func (s *KeeperTestSuite) TestListDateRequestWasm() {
	s.SetupTest()
	mockWasm1 := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	mockWasm2 := &wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}

	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm1)
	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm2)
	result := s.wasmStorageKeeper.ListDataRequestWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Equal(hex.EncodeToString(mockWasm1.Hash)+","+mockWasm1.WasmType.String(), result[0])
	s.Assert().Equal(hex.EncodeToString(mockWasm2.Hash)+","+mockWasm2.WasmType.String(), result[1])
}

func (s *KeeperTestSuite) TestListOverlayWasm() {
	s.SetupTest()
	mockWasm1 := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	mockWasm2 := &wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}

	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasm1)
	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasm2)
	result := s.wasmStorageKeeper.ListOverlayWasms(s.ctx)
	s.Assert().Equal(2, len(result))
	s.Assert().Equal(hex.EncodeToString(mockWasm1.Hash)+","+mockWasm1.WasmType.String(), result[0])
	s.Assert().Equal(hex.EncodeToString(mockWasm2.Hash)+","+mockWasm2.WasmType.String(), result[1])
}

func (s *KeeperTestSuite) TestGetAllWasms() {
	s.SetupTest()
	mockWasm1 := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}
	mockWasm2 := &wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeDataRequest,
	}

	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm1)
	s.wasmStorageKeeper.SetDataRequestWasm(s.ctx, mockWasm2)

	mockWasmO1 := &wasmstoragetypes.Wasm{
		Hash:     mockedByteArray,
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}
	mockWasmO2 := &wasmstoragetypes.Wasm{
		Hash:     append(mockedByteArray, 2),
		Bytecode: mockedByteArray,
		WasmType: wasmstoragetypes.WasmTypeRelayer,
	}

	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasmO1)
	s.wasmStorageKeeper.SetOverlayWasm(s.ctx, mockWasmO2)

	result := s.wasmStorageKeeper.GetAllWasms(s.ctx)
	s.Assert().Equal(4, len(result))
	s.Assert().Equal(*mockWasm1, result[0])
	s.Assert().Equal(*mockWasm2, result[1])
	s.Assert().Equal(*mockWasmO1, result[2])
	s.Assert().Equal(*mockWasmO2, result[3])
}
