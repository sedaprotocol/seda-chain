package types

import "cosmossdk.io/collections"

const (
	ModuleName = "pubkey"
	StoreKey
)

var (
	PubKeysPrefix        = collections.NewPrefix(0)
	ProvingSchemesPrefix = collections.NewPrefix(1)
	ParamsPrefix         = collections.NewPrefix(2)
)
