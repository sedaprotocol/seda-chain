package types

import "cosmossdk.io/collections"

const (
	ModuleName = "sophon"
	StoreKey
)

var (
	ParamsKey         = collections.NewPrefix(0)
	SophonIDKey       = collections.NewPrefix(1)
	SophonInfoKey     = collections.NewPrefix(2)
	SophonUserKey     = collections.NewPrefix(3)
	SophonTransferKey = collections.NewPrefix(4)
)
