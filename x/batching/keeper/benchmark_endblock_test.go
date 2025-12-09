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

	var lastBatchNum uint64
	for range numBatches {
		f.AddBlock()

		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(b, 1)[0])
		require.NoError(b, err)
		batch, dataEntries, valEntries, err := f.batchingKeeper.ConstructBatch(f.Context())
		require.NoError(b, err)
		lastBatchNum, err = f.batchingKeeper.SetNewBatch(f.Context(), batch, dataEntries, valEntries)
		require.NoError(b, err)
	}

	for b.Loop() {
		_, err := f.batchingKeeper.BatchPruneBatches(f.Context(), numBatchesToKeep, maxBatchPrunePerBlock, lastBatchNum)
		require.NoError(b, err)
	}
}
