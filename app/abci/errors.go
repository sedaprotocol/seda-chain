package abci

import "cosmossdk.io/errors"

const ModuleName = "vote_extension"

var (
	ErrNoBatchForCurrentHeight = errors.Register(ModuleName, 2, "no batch found for current height")
)
