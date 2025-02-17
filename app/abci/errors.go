package abci

import "cosmossdk.io/errors"

const ModuleName = "vote_extension"

var (
	ErrNoBatchForCurrentHeight      = errors.Register(ModuleName, 2, "no batch found for current height")
	ErrVoteExtensionTooLong         = errors.Register(ModuleName, 3, "vote extension exceeds max length")
	ErrVoteExtensionTooShort        = errors.Register(ModuleName, 4, "vote extension is too short")
	ErrVoteExtensionInjectionTooBig = errors.Register(ModuleName, 5, "injected vote extensions are too big")
	ErrInvalidBatchSignature        = errors.Register(ModuleName, 6, "batch signature is invalid")
	ErrUnexpectedBatchSignature     = errors.Register(ModuleName, 7, "batch signature should be empty")
)
