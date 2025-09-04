package types

import "cosmossdk.io/errors"

var (
	ErrFastClientAlreadyExists = errors.Register(ModuleName, 0, "fast client already exists")
	ErrUserAlreadyExists       = errors.Register(ModuleName, 1, "user already exists")
	ErrInsufficientCredits     = errors.Register(ModuleName, 2, "insufficient credits")
	ErrInsufficientBalance     = errors.Register(ModuleName, 3, "insufficient balance")
)
