package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "batching"

	// StoreKey defines the primary module store key.
	StoreKey = "batching"
)

var (
	DataResultsPrefix     = collections.NewPrefix(0)
	CurrentBatchNumberKey = collections.NewPrefix(1)
	BatchesKeyPrefix      = collections.NewPrefix(2)
	VotesKeyPrefix        = collections.NewPrefix(3)
	ParamsKey             = collections.NewPrefix(4)
)
