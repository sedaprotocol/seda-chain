package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "wasm-storage"

	// StoreKey defines the primary module store key
	StoreKey = "storage"

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// KeyParams is the store key for the parameters.
	KeyParams = "params"
)

var (
	// KeyPrefixDataRequest defines prefix to store Data Request Wasm binaries.
	KeyPrefixDataRequest = []byte{0x00}

	// KeyPrefixOverlay defines prefix to store Overlay Wasm binaries.
	KeyPrefixOverlay = []byte{0x01}

	// KeyPrefixDataRequestQueue defines prefix to store the queue that contains
	// the hashes of Data Request Wasm binaries.
	KeyPrefixDataRequestQueue = []byte{0x02}

	// KeyPrefixProxyContractRegistry defines prefix to store address of
	// Proxy Contract.
	KeyPrefixProxyContractRegistry = []byte{0x03}
)

func GetDataRequestWasmKey(hash []byte) []byte {
	return append(KeyPrefixDataRequest, hash...)
}

func GetOverlayWasmKey(hash []byte) []byte {
	return append(KeyPrefixOverlay, hash...)
}

// GetDataRequestTimeKey gets the key for an item in Data Request Queue. This key
// is the timestamp of when the Data Request Wasm was stored.
func GetDataRequestTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(KeyPrefixDataRequestQueue, bz...)
}
