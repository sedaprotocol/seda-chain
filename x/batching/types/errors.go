package types

import "cosmossdk.io/errors"

var (
	ErrBatchingHasNotStarted = errors.Register("batching", 2, "batching has not begun - there is no batch in the store")
	ErrInvalidBatchNumber    = errors.Register("batching", 3, "invalid batch number")
	ErrBatchAlreadyExists    = errors.Register("batching", 4, "batch already exists at the given block height")
	ErrInvalidPublicKey      = errors.Register("batching", 5, "invalid public key")
	ErrNoBatchingUpdate      = errors.Register("batching", 6, "no change from previous data result and validator roots")
	ErrNoSignedBatches       = errors.Register("batching", 7, "there is no signed batch yet")
)
