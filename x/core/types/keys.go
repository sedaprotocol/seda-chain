package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowlistKey              = collections.NewPrefix(0)
	StakersKeyPrefix          = collections.NewPrefix(1)
	StakerIndexToKeyPrefix    = collections.NewPrefix(2)
	StakerKeyToIndexPrefix    = collections.NewPrefix(3)
	StakerCountKey            = collections.NewPrefix(4)
	DataRequestsKeyPrefix     = collections.NewPrefix(5)
	RevealBodiesKeyPrefix     = collections.NewPrefix(6)
	DataRequestIndexingPrefix = collections.NewPrefix(7)
	TimeoutQueueKeyPrefix     = collections.NewPrefix(8)
	ParamsKey                 = collections.NewPrefix(9)
	OwnerKey                  = collections.NewPrefix(10)
	PausedKey                 = collections.NewPrefix(11)
	PendingOwnerKey           = collections.NewPrefix(12)
	DataRequestCommittingKey  = collections.NewPrefix(13)
	DataRequestRevealingKey   = collections.NewPrefix(14)
	DataRequestTallyingKey    = collections.NewPrefix(15)
)
