package types

import "cosmossdk.io/errors"

var ErrAlreadyExists = errors.Register(ModuleName, 0, "sophon already exists")
