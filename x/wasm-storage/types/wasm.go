package types

import (
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

// NewWasm constructs a new Wasm object given bytecode and Wasm type.
// It panics if it fails to compute hash of bytecode.
func NewWasm(bytecode []byte, wasmType WasmType) *Wasm {
	hash := crypto.Keccak256(bytecode)
	if hash == nil {
		panic("failed to compute hash")
	}

	return &Wasm{
		Hash:     hash,
		Bytecode: bytecode,
		WasmType: wasmType,
	}
}

func WasmTypeFromString(s string) WasmType {
	switch strings.ToUpper(s) {
	case "DATA_REQUEST", "WASM_TYPE_DATA_REQUEST":
		return WasmTypeDataRequest
	case "TALLY", "WASM_TYPE_TALLY":
		return WasmTypeTally
	case "DATA_REQUEST_EXECUTOR", "WASM_TYPE_DATA_REQUEST_EXECUTOR":
		return WasmTypeDataRequestExecutor
	case "RELAYER", "WASM_TYPE_RELAYER":
		return WasmTypeRelayer
	default:
		return WasmTypeNil
	}
}
