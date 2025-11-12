package types

import (
	"bytes"
	"encoding/hex"
	fmt "fmt"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/collections"
)

// NewGenesisState constructs a GenesisState object.
func NewGenesisState(
	curBatchNum uint64,
	batches []Batch,
	batchData []BatchData,
	dataResults []GenesisDataResult,
	batchAssignments []BatchAssignment,
	params Params,
) GenesisState {
	return GenesisState{
		CurrentBatchNumber: curBatchNum,
		Batches:            batches,
		BatchData:          batchData,
		DataResults:        dataResults,
		BatchAssignments:   batchAssignments,
		Params:             params,
	}
}

// DefaultGenesisState creates a default GenesisState object.
func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(collections.DefaultSequenceStart, nil, nil, nil, nil, DefaultParams())
	return &state
}

// ValidateGenesis validates batching genesis data.
func ValidateGenesis(gs GenesisState) error {
	for _, batch := range gs.Batches {
		if batch.BatchNumber > gs.CurrentBatchNumber {
			return fmt.Errorf("batch number %d should not exceed current batch number %d", batch.BatchNumber, gs.CurrentBatchNumber)
		}

		var provingMetaDataHash []byte
		if len(batch.ProvingMetadata) == 0 {
			provingMetaDataHash = make([]byte, 32)
		} else {
			hasher := sha3.NewLegacyKeccak256()
			hasher.Write(batch.ProvingMetadata)
			provingMetaDataHash = hasher.Sum(nil)
		}
		valRoot, err := hex.DecodeString(batch.ValidatorRoot)
		if err != nil {
			return err
		}
		dataRoot, err := hex.DecodeString(batch.DataResultRoot)
		if err != nil {
			return err
		}

		expectedBatchID := ComputeBatchID(batch.BatchNumber, batch.BlockHeight, valRoot, dataRoot, provingMetaDataHash)
		if !bytes.Equal(batch.BatchId, expectedBatchID) {
			return fmt.Errorf("batch id %s does not match expected batch id %s", hex.EncodeToString(batch.BatchId), hex.EncodeToString(expectedBatchID))
		}
	}

	return gs.Params.Validate()
}
