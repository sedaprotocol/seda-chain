package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "batching"

	// StoreKey defines the primary module store key.
	StoreKey = "batching"
)

var (
	CurrentBatchNumberKey = collections.NewPrefix(0)
	BatchPrefix           = collections.NewPrefix(1)
	ParamsKey             = collections.NewPrefix(2)
)
