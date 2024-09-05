package types

import "cosmossdk.io/collections"

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(curBatchNum uint64, batches []Batch, params Params) GenesisState {
	return GenesisState{
		CurrentBatchNumber: curBatchNum,
		Batches:            batches,
		Params:             params,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(collections.DefaultSequenceStart, nil, DefaultParams())
	return &state
}

// ValidateGenesis validates batching genesis data.
func ValidateGenesis(gs GenesisState) error {
	return nil
}