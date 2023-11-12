package types

const (
	// ModuleName defines the module name
	ModuleName = "randomness"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var (
	// KeyPrefixSeed defines prefix to store the current block's seed.
	KeyPrefixSeed = []byte{0x00}
)
