package types

import "cosmossdk.io/errors"

var (
	ErrEmptyValue        = errors.Register("pubkey", 1, "empty value")
	ErrInvalidType       = errors.Register("pubkey", 2, "invalid type")
	ErrValidatorNotFound = errors.Register("pubkey", 3, "validator not found")
)
