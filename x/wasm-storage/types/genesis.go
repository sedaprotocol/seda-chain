package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(params Params, drWasms []DataRequestWasm, execWasms []ExecutorWasm, coreAddr string) GenesisState {
	return GenesisState{
		Params:               params,
		DataRequestWasms:     drWasms,
		ExecutorWasms:        execWasms,
		CoreContractRegistry: coreAddr,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), nil, nil, "")
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
