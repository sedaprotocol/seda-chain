package types

import "cosmossdk.io/errors"

var (
	ErrInvalidFilterType   = errors.Register("tally", 2, "invalid filter type")
	ErrFilterInputTooShort = errors.Register("tally", 3, "filter input length too short")
	ErrInvalidPathLen      = errors.Register("tally", 4, "invalid JSON path length")
	ErrEmptyReveals        = errors.Register("tally", 5, "no reveals given")
	ErrCorruptReveals      = errors.Register("tally", 6, "> 1/3 of reveals are corrupted")
	ErrNoConsensus         = errors.Register("tally", 7, "> 1/3 of reveals do not agree on reveal data")
	ErrNoBasicConsensus    = errors.Register("tally", 8, "> 1/3 of reveals do not agree on (exit_code, proxy_pub_keys)")
	ErrInvalidNumberType   = errors.Register("tally", 9, "invalid number type specified")
	ErrFilterUnexpected    = errors.Register("tally", 10, "unexpected error occurred in filter")
	ErrInvalidSaltLength   = errors.Register("tally", 11, "salt should be 32-byte long")
)
