package types

import "cosmossdk.io/collections"

const (
	ModuleName = "core"
	StoreKey   = ModuleName
)

var (
	AllowListPrefix = collections.NewPrefix(0)
	ParamsPrefix    = collections.NewPrefix(1)
)
