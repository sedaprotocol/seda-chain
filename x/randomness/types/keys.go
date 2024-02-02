package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "randomness"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var SeedPrefix = collections.NewPrefix(0)
var ValidatorVRFPrefix = collections.NewPrefix(1)
