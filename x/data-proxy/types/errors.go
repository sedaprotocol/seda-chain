package types

import "cosmossdk.io/errors"

var (
	ErrEmptyValue       = errors.Register(ModuleName, 1, "empty value")
	ErrInvalidParam     = errors.Register(ModuleName, 2, "invalid parameter")
	ErrAlreadyExists    = errors.Register(ModuleName, 3, "already exists")
	ErrInvalidSignature = errors.Register(ModuleName, 4, "invalid signature")
	ErrInvalidAddress   = errors.Register(ModuleName, 5, "invalid address")
	ErrInvalidHex       = errors.Register(ModuleName, 6, "invalid hex string")
	ErrInvalidDelay     = errors.Register(ModuleName, 7, "invalid update delay")
	ErrUnknownDataProxy = errors.Register(ModuleName, 8, "unknown public key")
	ErrUnauthorized     = errors.Register(ModuleName, 9, "message not signed by admin")
	ErrEmptyUpdate      = errors.Register(ModuleName, 10, "nothing to update")
)
