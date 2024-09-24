package types

import "cosmossdk.io/errors"

var (
	ErrInvalidValidatorSetTrimPercent = errors.Register(ModuleName, 2, "validator set trim percent must be between 0 and 100")
)
