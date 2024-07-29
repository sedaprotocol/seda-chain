package types

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(ids []string) GenesisState {
	return GenesisState{
		KeyId: ids,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(nil)
	return &state
}

// ValidateGenesis validates wasm-storage genesis data.
func ValidateGenesis(gs GenesisState) error {
	return nil
}
