package types

import "cosmossdk.io/errors"

var (
	ErrInvalidParam          = errors.Register("wasm-storage", 1, "invalid param")
	ErrAlreadyExists         = errors.Register("wasm-storage", 2, "already exists")
	ErrInvalidFilterInputLen = errors.Register("wasm-storage", 3, "invalid filter input length")
	ErrInvalidPathLen        = errors.Register("wasm-storage", 4, "invalid JSON path length")
)
