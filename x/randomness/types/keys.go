package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

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

// KeyPrefixValidatorVRF defines prefix to store the validator VRF object.
var KeyPrefixValidatorVRF = []byte{0x01}

// GetValidatorVRFKey gets the key for the validator VRF object.
func GetValidatorVRFKey(consensusAddr sdk.ConsAddress) []byte {
	return append(KeyPrefixValidatorVRF, address.MustLengthPrefix(consensusAddr)...)
}
