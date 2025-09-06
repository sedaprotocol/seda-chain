package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowlistKey = collections.NewPrefix(0)
	StakersKey   = collections.NewPrefix(1)
	ParamsKey    = collections.NewPrefix(2)
)
