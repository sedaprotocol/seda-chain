package types

import "cosmossdk.io/collections"

const (
	ModuleName = "fast"
	StoreKey
)

var (
	ParamsKey             = collections.NewPrefix(0)
	FastClientIDKey       = collections.NewPrefix(1)
	FastClientKey         = collections.NewPrefix(2)
	FastUserKey           = collections.NewPrefix(3)
	FastClientTransferKey = collections.NewPrefix(4)
)
