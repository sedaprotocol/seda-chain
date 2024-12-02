package types

import "cosmossdk.io/collections"

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(
	curBatchNum uint64,
	batches []Batch,
	batchData []BatchData,
	dataResults []GenesisDataResult,
	batchAssignments []BatchAssignment,
) GenesisState {
	return GenesisState{
		CurrentBatchNumber: curBatchNum,
		Batches:            batches,
		BatchData:          batchData,
		DataResults:        dataResults,
		BatchAssignments:   batchAssignments,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(collections.DefaultSequenceStart, nil, nil, nil, nil)
	return &state
}

// ValidateGenesis validates batching genesis data.
func ValidateGenesis(_ GenesisState) error {
	return nil
}
