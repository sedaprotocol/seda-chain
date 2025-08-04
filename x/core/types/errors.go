package types

import (
	"cosmossdk.io/errors"
)

var (
	ErrAlreadyAllowlisted = errors.Register("core", 2, "public key already exists in allowlist")
	ErrNotAllowlisted     = errors.Register("core", 3, "public key is not in allowlist")
	ErrInsufficientStake  = errors.Register("core", 4, "stake amount is insufficient")
	ErrInvalidStakeProof  = errors.Register("core", 5, "invalid stake proof")
)
