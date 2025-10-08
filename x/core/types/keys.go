package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowlistKey           = collections.NewPrefix(0)
	StakersKeyPrefix       = collections.NewPrefix(1)
	StakerIndexToKeyPrefix = collections.NewPrefix(2)
	StakerKeyToIndexPrefix = collections.NewPrefix(3)
	StakerCountKey         = collections.NewPrefix(4)

	DataRequestsKeyPrefix         = collections.NewPrefix(5)
	DataRequestIndexingPrefix     = collections.NewPrefix(6)
	DataRequestCommittingCountKey = collections.NewPrefix(7)
	DataRequestRevealingCountKey  = collections.NewPrefix(8)
	DataRequestTallyingCountKey   = collections.NewPrefix(9)

	CommitsPrefix         = collections.NewPrefix(10)
	RevealersPrefix       = collections.NewPrefix(11)
	RevealBodiesKeyPrefix = collections.NewPrefix(12)
	TimeoutQueueKeyPrefix = collections.NewPrefix(13)

	ParamsKey       = collections.NewPrefix(14)
	OwnerKey        = collections.NewPrefix(15)
	PausedKey       = collections.NewPrefix(16)
	PendingOwnerKey = collections.NewPrefix(17)
)
