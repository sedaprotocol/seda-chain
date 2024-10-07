package types

import "cosmossdk.io/errors"

var (
	ErrBatchingHasNotStarted = errors.Register("batching", 2, "batching has not begun - there is no batch in the store")
	ErrInvalidBatchNumber    = errors.Register("batching", 3, "invalid batch number")
	ErrBatchAlreadyExists    = errors.Register("batching", 4, "batch already exists at the given block height")
)
