package types

import "cosmossdk.io/errors"

var (
	ErrInvalidFilterType   = errors.Register("tally", 2, "invalid filter type")
	ErrFilterInputTooShort = errors.Register("tally", 3, "filter input length too short")
	ErrInvalidPathLen      = errors.Register("tally", 4, "invalid JSON path length")
	ErrEmptyReveals        = errors.Register("tally", 5, "no reveals given")
	ErrCorruptReveals      = errors.Register("tally", 6, "more than 1/3 of the reveals are corrupted")
	ErrNoConsensus         = errors.Register("tally", 7, "1/3 or more of the reveals are not in consensus range")
	ErrInvalidNumberType   = errors.Register("tally", 8, "invalid number type specified")
	ErrFilterUnexpected    = errors.Register("tally", 9, "unexpected error occurred in filter")
)
