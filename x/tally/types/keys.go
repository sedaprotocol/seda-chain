package types

import "cosmossdk.io/collections"

const (
	ModuleName = "tally"
	StoreKey   = ModuleName
)

var ParamsPrefix = collections.NewPrefix(0)
