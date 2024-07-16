package types

import "cosmossdk.io/errors"

var (
	ErrInvalidParam          = errors.Register("wasm-storage", 1, "invalid param")
	ErrAlreadyExists         = errors.Register("wasm-storage", 2, "already exists")
	ErrInvalidFilterType     = errors.Register("wasm-storage", 3, "invalid filter type")
	ErrInvalidFilterInputLen = errors.Register("wasm-storage", 4, "invalid filter input length")
	ErrInvalidPathLen        = errors.Register("wasm-storage", 5, "invalid JSON path length")
	ErrEmptyReveals          = errors.Register("wasm-storage", 6, "no reveals given")
	ErrCorruptReveals        = errors.Register("wasm-storage", 7, "more than 1/3 of the reveals are corrupted")
	ErrNoConsensus           = errors.Register("wasm-storage", 8, "less than 2/3 of the reveals in consensus range")
	ErrInvalidNumberType     = errors.Register("wasm-storage", 9, "invalid number type specified")
)
