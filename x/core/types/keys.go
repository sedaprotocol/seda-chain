package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowlistKey          = collections.NewPrefix(0)
	StakersKeyPrefix      = collections.NewPrefix(1)
	ParamsKey             = collections.NewPrefix(2)
	DataRequestsKeyPrefix = collections.NewPrefix(3)
	CommittingKeyPrefix   = collections.NewPrefix(4)
	RevealingKeyPrefix    = collections.NewPrefix(5)
	TallyingKeyPrefix     = collections.NewPrefix(6)
	TimeoutQueueKeyPrefix = collections.NewPrefix(7)
)
