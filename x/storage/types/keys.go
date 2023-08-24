package types

const (
	// ModuleName defines the module name
	ModuleName = "storage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

var DataStoreKeyPrefix = []byte{0x00}

func DataStoreKey(hash []byte) []byte {
	return append(DataStoreKeyPrefix, hash...)
}
