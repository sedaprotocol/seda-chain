package types

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates core genesis data.
func ValidateGenesis(state GenesisState) error {
	return state.Params.Validate()
}
