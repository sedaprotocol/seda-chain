package types

import (
	fmt "fmt"
)

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
