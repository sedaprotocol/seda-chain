package types

import "cosmossdk.io/errors"

var (
	ErrNilValue     = errors.Register("pkr", 1, "nil value")
	ErrInvalidType  = errors.Register("pkr", 2, "invalid type")
	ErrInvalidInput = errors.Register("pkr", 3, "invalid input")
)
