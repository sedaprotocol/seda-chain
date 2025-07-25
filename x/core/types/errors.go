package types

import (
	"cosmossdk.io/errors"
)

var (
	ErrAlreadyAllowlisted = errors.Register("core", 2, "public key already exists in allowlist")
)
