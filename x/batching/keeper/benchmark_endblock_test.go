package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkBatchPruning(b *testing.B) {
	f := initFixture(b)
	f.addBatchSigningValidators(b, 100)

	numBatches := 2000
	numBatchesToKeep := uint64(1000)
	maxBatchPrunePerBlock := uint64(100)

	for range numBatches {
		f.AddBlock()

		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(b, 1)[0])
		require.NoError(b, err)
		batch, dataEntries, valEntries, err := f.batchingKeeper.ConstructBatch(f.Context())
		require.NoError(b, err)
		err = f.batchingKeeper.SetNewBatch(f.Context(), batch, dataEntries, valEntries)
		require.NoError(b, err)
	}

	for b.Loop() {
		_, err := f.batchingKeeper.PruneBatches(f.Context(), numBatchesToKeep, maxBatchPrunePerBlock)
		require.NoError(b, err)
	}
}

func BenchmarkDataResultPruning(b *testing.B) {
	f := initFixture(b)

	maxDataResultsToCheckForPrune := uint64(100)

	// Create 10 data results for each of 1000 batches
	for i := range uint64(100) {
		f.AddBlock()

		dataResults := generateDataResults(b, 10)
		for _, dataResult := range dataResults {
			err := f.batchingKeeper.SetDataResultForBatching(f.Context(), dataResult)
			require.NoError(b, err)
			err = f.batchingKeeper.MarkDataResultAsBatched(f.Context(), dataResult, i)
			require.NoError(b, err)
		}
	}

	for b.Loop() {
		f.AddBlock()
		f.SetRandomLastCommitHash()

		err := f.batchingKeeper.PruneDataResults(f.Context(), maxDataResultsToCheckForPrune, 2000)
		require.NoError(b, err)
	}
}
