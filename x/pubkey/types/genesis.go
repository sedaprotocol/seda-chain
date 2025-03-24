package types

import (
	fmt "fmt"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
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
	// Ensure secp256k1 proving scheme exists to prevent panic in batching end blocker.
	found := false
	for _, scheme := range data.ProvingSchemes {
		if scheme.Index == uint32(sedatypes.SEDAKeyIndexSecp256k1) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("secp256k1 proving scheme is required")
	}

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
