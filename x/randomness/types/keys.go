package types

const (
	// ModuleName defines the module name
	ModuleName = "randomness"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KeyPrefixSeed defines prefix to store the current block's seed.
var KeyPrefixSeed = []byte{0x00}
