package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/batching/keeper"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// TestMigration tests the migration of unbatched legacy data results to the new collection.
func TestMigration(t *testing.T) {
	f := initFixture(t)

	// Create 5 batches with 10 data results each before mock upgrade.
	// We simulate the chain before the upgrade by using legacy functions.
	for range 5 {
		f.AddBlock()

		dataResults := generateDataResults(t, 10)
		for _, dataResult := range dataResults {
			err := f.batchingKeeper.LegacySetDataResultForBatching(f.Context(), dataResult)
			require.NoError(t, err)
		}
		batch, dataEntries, valEntries, err := f.batchingKeeper.LegacyConstructBatch(f.Context())
		require.NoError(t, err)
		_, err = f.batchingKeeper.SetNewBatch(f.Context(), batch, dataEntries, valEntries)
		require.NoError(t, err)
	}

	// Execute the migration.
	migrator := keeper.NewMigrator(f.batchingKeeper)
	err := migrator.Migrate1to2(f.Context())
	require.NoError(t, err)

	// Check that batchNumberAtUpgrade and hasPruningCaughtUp have been set.
	batchNumberAtUpgrade, err := f.batchingKeeper.GetBatchNumberAtUpgrade(f.Context())
	require.NoError(t, err)
	require.Equal(t, uint64(4), batchNumberAtUpgrade)

	hasPruningCaughtUp, err := f.batchingKeeper.HasPruningCaughtUp(f.Context())
	require.NoError(t, err)
	require.False(t, hasPruningCaughtUp)

	params, err := f.batchingKeeper.GetParams(f.Context())
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), params)
}
