package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(params Params, wasms []Wasm, proxyAddr string) GenesisState {
	return GenesisState{
		Params:                params,
		Wasms:                 wasms,
		ProxyContractRegistry: proxyAddr,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), nil, "")
	return &state
}

// ValidateGenesis validates wasm-storage genesis data.
func ValidateGenesis(gs GenesisState) error {
	if gs.ProxyContractRegistry != "" {
		_, err := sdk.AccAddressFromBech32(gs.ProxyContractRegistry)
		if err != nil {
			return fmt.Errorf("invalid Proxy contract address %w", err)
		}
	}
	return gs.Params.ValidateBasic()
}
