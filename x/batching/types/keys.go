package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "batching"

	// StoreKey defines the primary module store key.
	StoreKey = "batching"
)

var (
	LegacyDataResultsPrefix        = collections.NewPrefix(0)
	BatchAssignmentsPrefix         = collections.NewPrefix(1)
	CurrentBatchNumberKey          = collections.NewPrefix(2)
	BatchesKeyPrefix               = collections.NewPrefix(3)
	BatchNumberKeyPrefix           = collections.NewPrefix(4)
	ValidatorTreeEntriesKeyPrefix  = collections.NewPrefix(5)
	DataResultTreeEntriesKeyPrefix = collections.NewPrefix(6)
	BatchSignaturesKeyPrefix       = collections.NewPrefix(7)
	ParamsKey                      = collections.NewPrefix(8)
	DataResultsPrefix              = collections.NewPrefix(9)
	BatchDataResultsPrefix         = collections.NewPrefix(10)
	BatchNumberAtUpgradeKey        = collections.NewPrefix(11)
	HasPruningCaughtUpKey          = collections.NewPrefix(12)
)
