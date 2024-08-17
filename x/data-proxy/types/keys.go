package types

import "cosmossdk.io/collections"

const (
	ModuleName = "data-proxy"
	StoreKey
)

var (
	DataProxyConfigPrefix = collections.NewPrefix(0)
	ParamsPrefix          = collections.NewPrefix(1)
	FeeUpdatesPrefix      = collections.NewPrefix(2)
)
