package types

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(wasms []Wasm) GenesisState {
	return GenesisState{
		Wasms: wasms,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(nil)
	return &state
}

// ValidateGenesis validates wasm-storage genesis data.
func ValidateGenesis(genesisState GenesisState) error {
	return nil
}
