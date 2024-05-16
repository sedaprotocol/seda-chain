package types

import "cosmossdk.io/errors"

var (
	ErrInvalidParam  = errors.Register("wasm-storage", 1, "invalid param")
	ErrAlreadyExists = errors.Register("wasm-storage", 2, "already exists")
)
