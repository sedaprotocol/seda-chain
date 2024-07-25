package types

import "cosmossdk.io/errors"

var (
	ErrInvalidFilterType     = errors.Register("tally", 1, "invalid filter type")
	ErrInvalidFilterInputLen = errors.Register("tally", 2, "invalid filter input length")
	ErrInvalidPathLen        = errors.Register("tally", 3, "invalid JSON path length")
	ErrEmptyReveals          = errors.Register("tally", 4, "no reveals given")
	ErrCorruptReveals        = errors.Register("tally", 5, "more than 1/3 of the reveals are corrupted")
	ErrNoConsensus           = errors.Register("tally", 6, "1/3 or more of the reveals are not in consensus range")
	ErrInvalidNumberType     = errors.Register("tally", 7, "invalid number type specified")
	ErrFilterUnexpected      = errors.Register("tally", 8, "unexpected error occurred in filter")
)
