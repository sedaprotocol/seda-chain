package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "batching"

	// StoreKey defines the primary module store key.
	StoreKey = "batching"
)

var (
	DataResultsPrefix        = collections.NewPrefix(0)
	BatchAssignmentsPrefix   = collections.NewPrefix(1)
	CurrentBatchNumberKey    = collections.NewPrefix(2)
	BatchesKeyPrefix         = collections.NewPrefix(3)
	BatchNumberKeyPrefix     = collections.NewPrefix(4)
	TreeEntriesKeyPrefix     = collections.NewPrefix(5)
	BatchSignaturesKeyPrefix = collections.NewPrefix(6)
	ParamsKey                = collections.NewPrefix(7)
)
