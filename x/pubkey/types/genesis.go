package types

import (
	fmt "fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces.
func (g GenesisState) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for i := range g.ValidatorPubKeys {
		if err := g.ValidatorPubKeys[i].UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces.
func (v ValidatorPubKeys) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range v.IndexedPubKeys {
		var pubKey cryptotypes.PubKey
		err := unpacker.UnpackAny(v.IndexedPubKeys[i].PubKey, &pubKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewGenesisState(valPubKeys []ValidatorPubKeys) GenesisState {
	return GenesisState{
		ValidatorPubKeys: valPubKeys,
	}
}

func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(nil)
	return &state
}

func ValidateGenesis(data GenesisState) error {
	for _, val := range data.ValidatorPubKeys {
		if val.ValidatorAddr == "" {
			return fmt.Errorf("empty validator address")
		}
		for _, pk := range val.IndexedPubKeys {
			if pk.PubKey == nil {
				return fmt.Errorf("empty public key at index %d validator %s", pk.Index, val.ValidatorAddr)
			}
		}
	}
	return nil
}
