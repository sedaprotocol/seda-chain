package types

const DefaultSeed = "sedouards"

// ValidateGenesis ensures validity of given randomness genesis state.
func ValidateGenesis(data GenesisState) error {
	if data.Seed == "" {
		return ErrEmptyRandomnessSeed
	}
	return nil
}

// DefaultGenesisState returns default state for randomness module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Seed: DefaultSeed,
	}
}
