package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowlistKey              = collections.NewPrefix(0)
	StakersKeyPrefix          = collections.NewPrefix(1)
	DataRequestsKeyPrefix     = collections.NewPrefix(2)
	RevealBodiesKeyPrefix     = collections.NewPrefix(3)
	DataRequestIndexingPrefix = collections.NewPrefix(4)
	TimeoutQueueKeyPrefix     = collections.NewPrefix(5)
	ParamsKey                 = collections.NewPrefix(6)
	OwnerKey                  = collections.NewPrefix(7)
	PausedKey                 = collections.NewPrefix(8)
	PendingOwnerKey           = collections.NewPrefix(9)
	DataRequestCommittingKey  = collections.NewPrefix(10)
	DataRequestRevealingKey   = collections.NewPrefix(11)
	DataRequestTallyingKey    = collections.NewPrefix(12)
)
