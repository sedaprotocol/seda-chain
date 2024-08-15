package types

func NewGenesisState() GenesisState {
	return GenesisState{}
}

func DefaultGenesisState() *GenesisState {
	state := NewGenesisState()
	return &state
}

func ValidateGenesis(_ GenesisState) error {
	return nil
}
