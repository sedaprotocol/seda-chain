package types

import (
	fmt "fmt"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ProvingSchemes: []ProvingScheme{
			{
				Index:     0, // TODO resolve import cycle for uint32(utils.SEDAKeyIndexSecp256k1),
				IsEnabled: false,
			},
		},
	}
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
