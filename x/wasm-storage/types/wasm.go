package types

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// MaxWasmSize is the maximum size of wasm bytecode.
	MaxWasmSize = 800 * 1024
	// MinWasmSize is the realistic minimum size of wasm bytecode.
	MinWasmSize = 20
)

func validateWasmSize(s []byte) error {
	if len(s) < MinWasmSize {
		return fmt.Errorf("wasm code must be larger than %d bytes", MinWasmSize)
	}
	if len(s) > MaxWasmSize {
		return fmt.Errorf("wasm code cannot be larger than %d bytes", MaxWasmSize)
	}
	return nil
}

// NewDataRequestWasm constructs a new DataRequestWasm object given
// bytecode. It panics if it fails to compute hash of bytecode.
func NewDataRequestWasm(bytecode []byte, addedAt time.Time, curBlock, ttl int64) DataRequestWasm {
	hash := crypto.Keccak256(bytecode)
	if hash == nil {
		panic("failed to compute hash")
	}
	var expHeight int64
	if ttl > 0 {
		expHeight = curBlock + ttl
	}
	return DataRequestWasm{
		Hash:             hash,
		Bytecode:         bytecode,
		AddedAt:          addedAt,
		ExpirationHeight: expHeight,
	}
}

// NewExecutorWasm constructs a new ExecutorWasm object given bytecode.
// It panics if it fails to compute hash of bytecode.
func NewExecutorWasm(bytecode []byte, addedAt time.Time) ExecutorWasm {
	hash := crypto.Keccak256(bytecode)
	if hash == nil {
		panic("failed to compute hash")
	}
	return ExecutorWasm{
		Hash:     hash,
		Bytecode: bytecode,
		AddedAt:  addedAt,
	}
}
