package types

import "cosmossdk.io/errors"

var (
	ErrAlreadyExists     = errors.Register(ModuleName, 2, "data proxy already exists")
	ErrInvalidSignature  = errors.Register(ModuleName, 3, "invalid signature")
	ErrInvalidDelay      = errors.Register(ModuleName, 4, "invalid update delay")
	ErrEmptyUpdate       = errors.Register(ModuleName, 5, "nothing to update")
	ErrMaxUpdatesReached = errors.Register(ModuleName, 6, "max fee updates reached for block")
)
