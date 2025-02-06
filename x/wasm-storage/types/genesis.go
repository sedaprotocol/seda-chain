package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(params Params, programs []OracleProgram, coreAddr string) GenesisState {
	return GenesisState{
		Params:               params,
		OraclePrograms:       programs,
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
			return err
		}
	}
	return gs.Params.Validate()
}
