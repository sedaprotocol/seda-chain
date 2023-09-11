package types

import (
	fmt "fmt"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

// MaxWasmSize is the maximum size of Wasm bytecode.
const MaxWasmSize = 800 * 1024

func validateWasmCode(s []byte) error {
	if len(s) == 0 {
		return fmt.Errorf("empty Wasm code")
	}
	if len(s) > MaxWasmSize {
		return fmt.Errorf("Wasm code cannot be longer than %d bytes", MaxWasmSize)
	}
	return nil
}

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
	case "DATA-REQUEST":
		return WasmTypeDataRequest
	case "TALLY":
		return WasmTypeTally
	case "DATA-REQUEST-EXECUTOR":
		return WasmTypeDataRequestExecutor
	case "RELAYER":
		return WasmTypeRelayer
	default:
		return WasmTypeNil
	}
}
