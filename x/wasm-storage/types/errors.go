package types

import "cosmossdk.io/errors"

var (
	ErrInvalidParam          = errors.Register(ModuleName, 2, "invalid param")
	ErrWasmAlreadyExists     = errors.Register(ModuleName, 3, "wasm with the same hash already exists")
	ErrWasmTooLarge          = errors.Register(ModuleName, 4, "wasm size is too large")
	ErrWasmTooSmall          = errors.Register(ModuleName, 5, "wasm size is too small")
	ErrInvalidHexWasmHash    = errors.Register(ModuleName, 6, "invalid hex-encoded wasm hash")
	ErrWasmNotGzipCompressed = errors.Register(ModuleName, 7, "wasm is not gzip compressed")
	ErrInvalidAuthority      = errors.Register(ModuleName, 8, "invalid authority")
)
