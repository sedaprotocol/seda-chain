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
	RevealsKeyPrefix      = collections.NewPrefix(4)
	CommittingKeyPrefix   = collections.NewPrefix(5)
	RevealingKeyPrefix    = collections.NewPrefix(6)
	TallyingKeyPrefix     = collections.NewPrefix(7)
	TimeoutQueueKeyPrefix = collections.NewPrefix(8)
)
