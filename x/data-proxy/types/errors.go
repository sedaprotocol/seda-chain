package types

import "cosmossdk.io/errors"

var (
	ErrEmptyValue   = errors.Register(ModuleName, 1, "empty value")
	ErrInvalidParam = errors.Register(ModuleName, 2, "invalid parameter")
)
