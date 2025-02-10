package types

import (
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// MaxWasmSize is the maximum size of gzipped wasm bytecode that we'll accept.
	MaxWasmSize = 1024 * 1024
	// MinWasmSize is the realistic minimum size of gzipped wasm bytecode.
	MinWasmSize = 20
)

func validateWasmSize(s []byte) error {
	if len(s) < MinWasmSize {
		return ErrWasmTooSmall.Wrapf("%d < %d", len(s), MinWasmSize)
	}
	if len(s) > MaxWasmSize {
		return ErrWasmTooLarge.Wrapf("%d > %d", len(s), MaxWasmSize)
	}
	return nil
}

// NewOracleProgram constructs a new OracleProgram object given
// bytecode. It panics if it fails to compute hash of bytecode.
func NewOracleProgram(bytecode []byte, addedAt time.Time) OracleProgram {
	hash := crypto.Keccak256(bytecode)
	if hash == nil {
		panic("failed to compute hash")
	}
	return OracleProgram{
		Hash:     hash,
		Bytecode: bytecode,
		AddedAt:  addedAt,
	}
}
