package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(params Params, wasms []Wasm, coreAddr string) GenesisState {
	return GenesisState{
		Params:               params,
		Wasms:                wasms,
		CoreContractRegistry: coreAddr,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), nil, "")
	return &state
}

// ValidateGenesis validates wasm-storage genesis data.
func ValidateGenesis(gs GenesisState) error {
	if gs.CoreContractRegistry != "" {
		_, err := sdk.AccAddressFromBech32(gs.CoreContractRegistry)
		if err != nil {
			return fmt.Errorf("invalid Core contract address %w", err)
		}
	}
	return gs.Params.ValidateBasic()
}
