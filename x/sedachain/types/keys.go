package types

var ParamsKey = []byte{0x00}

const (
	// ModuleName defines the module name
	ModuleName = "sedachain"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)
