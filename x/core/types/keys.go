package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowlistKey          = collections.NewPrefix(0)
	StakersKeyPrefix      = collections.NewPrefix(1)
	DataRequestsKeyPrefix = collections.NewPrefix(2)
	RevealBodiesKeyPrefix = collections.NewPrefix(3)
	CommittingKeyPrefix   = collections.NewPrefix(4)
	RevealingKeyPrefix    = collections.NewPrefix(5)
	TallyingKeyPrefix     = collections.NewPrefix(6)
	TimeoutQueueKeyPrefix = collections.NewPrefix(7)
	ParamsKey             = collections.NewPrefix(8)
)
