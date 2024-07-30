package types

import "cosmossdk.io/errors"

var (
	ErrEmptyValue        = errors.Register("pkr", 1, "empty value")
	ErrInvalidType       = errors.Register("pkr", 2, "invalid type")
	ErrValidatorNotFound = errors.Register("pkr", 3, "validator not found")
)
