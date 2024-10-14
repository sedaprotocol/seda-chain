package abci

import "cosmossdk.io/errors"

const ModuleName = "vote_extension"

var (
	ErrNoBatchForCurrentHeight      = errors.Register(ModuleName, 2, "no batch found for current height")
	ErrInvalidVoteExtensionLength   = errors.Register(ModuleName, 3, "invalid vote extension length")
	ErrVoteExtensionInjectionTooBig = errors.Register(ModuleName, 4, "injected vote extensions are too big")
	ErrInvalidBatchSignature        = errors.Register(ModuleName, 5, "batch signature is invalid")
)
