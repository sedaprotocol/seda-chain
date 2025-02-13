package types

import (
	fmt "fmt"
)

const (
	DefaultActivationHeight = -1 // -1 indicates that activation is not in progress.
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		ProvingSchemes: []ProvingScheme{
			{
				Index:            0, // SEDA Key Index for Secp256k1
				IsActivated:      false,
				ActivationHeight: DefaultActivationHeight,
			},
		},
		Params: DefaultParams(),
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
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return nil
}
