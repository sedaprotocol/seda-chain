package types

import "cosmossdk.io/errors"

var (
	ErrSophonAlreadyExists = errors.Register(ModuleName, 0, "sophon already exists")
	ErrUserAlreadyExists   = errors.Register(ModuleName, 1, "user already exists")
	ErrInsufficientCredits = errors.Register(ModuleName, 2, "insufficient credits")
)
