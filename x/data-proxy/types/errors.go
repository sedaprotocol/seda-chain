package types

import "cosmossdk.io/errors"

var (
	ErrEmptyValue       = errors.Register(ModuleName, 1, "empty value")
	ErrInvalidParam     = errors.Register(ModuleName, 2, "invalid parameter")
	ErrAlreadyExists    = errors.Register(ModuleName, 3, "data proxy already exists")
	ErrInvalidSignature = errors.Register(ModuleName, 4, "invalid signature")
	ErrInvalidDelay     = errors.Register(ModuleName, 5, "invalid update delay")
	ErrEmptyUpdate      = errors.Register(ModuleName, 6, "nothing to update")
)
