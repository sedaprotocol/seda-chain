package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/batching/keeper"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func TestBatchPruneBatches(t *testing.T) {
	f := initFixture(t)

	f.addBatchSigningValidators(t, 10)

	numBatchesToKeep := uint64(75)
	maxBatchPrunePerBlock := uint64(150)

	err := f.batchingKeeper.SetParams(f.Context(), types.Params{
		MaxBatchPrunePerBlock: maxBatchPrunePerBlock,
	})
	require.NoError(t, err)

	// Should prune nothing.
	hasCaughtUp, err := f.batchingKeeper.BatchPruneBatches(f.Context(), numBatchesToKeep, maxBatchPrunePerBlock)
	require.NoError(t, err)
	require.False(t, hasCaughtUp)

	// Create 300 batches with random associated data.
	for range 300 {
		f.AddBlock()

		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(t, 1)[0])
		require.NoError(t, err)
		batch, dataEntries, valEntries, err := f.batchingKeeper.ConstructBatch(f.Context())
		require.NoError(t, err)
		_, err = f.batchingKeeper.SetNewBatch(f.Context(), batch, dataEntries, valEntries)
		require.NoError(t, err)
		err = f.batchingKeeper.SetBatchSigSecp256k1(f.Context(), batch.BatchNumber, valEntries[0].ValidatorAddress, generateRandomBytes(64))
		require.NoError(t, err)
	}

	batches, err := f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 300, len(batches))

	// Suppose an upgrade happens here and sets batchNumberAtUpgrade.
	err = f.batchingKeeper.SetBatchNumberAtUpgrade(f.Context())
	require.NoError(t, err)
	err = f.batchingKeeper.SetHasPruningCaughtUp(f.Context(), false)
	require.NoError(t, err)

	// Should prune first 150 batches (0-149)
	hasCaughtUp, err = f.batchingKeeper.BatchPruneBatches(f.Context(), numBatchesToKeep, maxBatchPrunePerBlock)
	require.NoError(t, err)
	require.False(t, hasCaughtUp)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 150, len(batches))
	require.Equal(t, uint64(150), batches[0].BatchNumber)
	require.Equal(t, uint64(299), batches[len(batches)-1].BatchNumber)

	for i := 0; i <= 149; i++ {
		f.checkNoBatchData(t, uint64(i))
	}
	for i := 150; i <= 299; i++ {
		f.checkBatchData(t, uint64(i), true)
	}

	// Should prune second 75 batches (150-224)
	hasCaughtUp, err = f.batchingKeeper.BatchPruneBatches(f.Context(), numBatchesToKeep, maxBatchPrunePerBlock)
	require.NoError(t, err)
	require.True(t, hasCaughtUp)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 75, len(batches))
	require.Equal(t, uint64(225), batches[0].BatchNumber)
	require.Equal(t, uint64(299), batches[len(batches)-1].BatchNumber)

	for i := 0; i <= 224; i++ {
		f.checkNoBatchData(t, uint64(i))
	}
	for i := 225; i <= 299; i++ {
		f.checkBatchData(t, uint64(i), true)
	}

	// Should prune nothing
	hasCaughtUp, err = f.batchingKeeper.BatchPruneBatches(f.Context(), numBatchesToKeep, maxBatchPrunePerBlock)
	require.NoError(t, err)
	require.True(t, hasCaughtUp)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 75, len(batches))
	require.Equal(t, uint64(225), batches[0].BatchNumber)
	require.Equal(t, uint64(299), batches[len(batches)-1].BatchNumber)

	for i := 0; i <= 224; i++ {
		f.checkNoBatchData(t, uint64(i))
	}
	for i := 225; i <= 299; i++ {
		f.checkBatchData(t, uint64(i), true)
	}
}

// TestLegacyDataResultPruning creates 1000 legacy data results and tests their
// pruning without creating batches.
func TestLegacyDataResultPruning(t *testing.T) {
	f := initFixture(t)

	err := f.pubKeyKeeper.SetProvingScheme(f.Context(), pubkeytypes.ProvingScheme{
		Index:       uint32(sedatypes.SEDAKeyIndexSecp256k1),
		IsActivated: true,
	})
	require.NoError(t, err)

	err = f.batchingKeeper.SetHasPruningCaughtUp(f.Context(), false)
	require.NoError(t, err)

	// Adjust the global variable for the test.
	original := keeper.NumBatchesToKeep
	defer func() {
		keeper.NumBatchesToKeep = original
	}()
	keeper.NumBatchesToKeep = 10

	err = f.batchingKeeper.SetParams(f.Context(), types.Params{MaxLegacyDataResultPrunePerBlock: 101})
	require.NoError(t, err)

	// Create 10 data results for each of 100 batches
	for i := range uint64(100) {
		dataResults := generateDataResults(t, 10)
		for _, dataResult := range dataResults {
			err := f.batchingKeeper.LegacySetDataResultForBatching(f.Context(), dataResult)
			require.NoError(t, err)
			err = f.batchingKeeper.LegacyMarkDataResultAsBatched(f.Context(), dataResult, i)
			require.NoError(t, err)
		}
	}

	dataResults, err := f.batchingKeeper.GetLegacyDataResults(f.Context(), true)
	require.NoError(t, err)
	require.Equal(t, 1000, len(dataResults))

	for _, dataResult := range dataResults {
		_, err := f.batchingKeeper.GetDataResult(f.Context(), dataResult.DrId, dataResult.DrBlockHeight)
		require.NoError(t, err)
		_, err = f.batchingKeeper.GetBatchAssignment(f.Context(), dataResult.DrId, dataResult.DrBlockHeight)
		require.NoError(t, err)
	}

	// Activate pruning of legacy data results.
	err = f.batchingKeeper.SetHasPruningCaughtUp(f.Context(), true)
	require.NoError(t, err)
	err = f.batchingKeeper.SetBatchNumberAtUpgrade(f.Context())
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		f.AddBlock()

		err = f.batchingKeeper.EndBlock(f.Context())
		require.NoError(t, err)

		res, err := f.batchingKeeper.GetLegacyDataResults(f.Context(), true)
		require.NoError(t, err)

		expectedCount := max(1000-101*(i+1), 0)
		require.Equal(t, expectedCount, len(res))
	}

	for _, dataResult := range dataResults {
		_, err := f.batchingKeeper.GetDataResult(f.Context(), dataResult.DrId, dataResult.DrBlockHeight)
		require.ErrorIs(t, err, collections.ErrNotFound)
		_, err = f.batchingKeeper.GetBatchAssignment(f.Context(), dataResult.DrId, dataResult.DrBlockHeight)
		require.ErrorIs(t, err, collections.ErrNotFound)
	}
}

// TestBasicPruning tests basic pruning with batch pruning disabled.
func TestBasicPruning(t *testing.T) {
	f := initFixture(t)

	f.addBatchSigningValidators(t, 10)

	err := f.pubKeyKeeper.SetProvingScheme(f.Context(), pubkeytypes.ProvingScheme{
		Index:       uint32(sedatypes.SEDAKeyIndexSecp256k1),
		IsActivated: true,
	})
	require.NoError(t, err)

	// Adjust the global variable for the test.
	original := keeper.NumBatchesToKeep
	defer func() {
		keeper.NumBatchesToKeep = original
	}()
	keeper.NumBatchesToKeep = 15

	params := types.Params{MaxBatchPrunePerBlock: 0} // Disable batch pruning
	err = f.batchingKeeper.SetParams(f.Context(), params)
	require.NoError(t, err)

	// Create 30 batches with random associated data.
	for range 30 {
		f.AddBlock()

		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(t, 1)[0])
		require.NoError(t, err)
		batch, dataEntries, valEntries, err := f.batchingKeeper.ConstructBatch(f.Context())
		require.NoError(t, err)
		_, err = f.batchingKeeper.SetNewBatch(f.Context(), batch, dataEntries, valEntries)
		require.NoError(t, err)
		err = f.batchingKeeper.SetBatchSigSecp256k1(f.Context(), batch.BatchNumber, valEntries[0].ValidatorAddress, generateRandomBytes(64))
		require.NoError(t, err)
	}

	batches, err := f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 30, len(batches))

	// Should not prune anything because no batch is created.
	f.AddBlock()
	err = f.batchingKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 30, len(batches))

	// Should create 31st batch Batch 30 and prune Batch 15.
	f.AddBlock()
	err = f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(t, 1)[0])
	require.NoError(t, err)
	err = f.batchingKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 30, len(batches))

	for i := 0; i < 15; i++ {
		f.checkBatchData(t, uint64(i), true)
	}
	f.checkNoBatchData(t, 15)
	for i := 16; i <= 30; i++ {
		f.checkBatchData(t, uint64(i), false)
	}
}

func TestNoBasicPruningUntilNumBatchesToKeepIsReached(t *testing.T) {
	f := initFixture(t)

	f.addBatchSigningValidators(t, 10)

	err := f.pubKeyKeeper.SetProvingScheme(f.Context(), pubkeytypes.ProvingScheme{
		Index:       uint32(sedatypes.SEDAKeyIndexSecp256k1),
		IsActivated: true,
	})
	require.NoError(t, err)

	// Adjust the global variable for the test.
	original := keeper.NumBatchesToKeep
	defer func() {
		keeper.NumBatchesToKeep = original
	}()
	keeper.NumBatchesToKeep = 11

	err = f.batchingKeeper.SetParams(f.Context(), types.Params{MaxBatchPrunePerBlock: 0})
	require.NoError(t, err)

	// Create 10 batches with random associated data.
	for range 10 {
		f.AddBlock()

		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(t, 1)[0])
		require.NoError(t, err)
		batch, dataEntries, valEntries, err := f.batchingKeeper.ConstructBatch(f.Context())
		require.NoError(t, err)
		_, err = f.batchingKeeper.SetNewBatch(f.Context(), batch, dataEntries, valEntries)
		require.NoError(t, err)
		err = f.batchingKeeper.SetBatchSigSecp256k1(f.Context(), batch.BatchNumber, valEntries[0].ValidatorAddress, generateRandomBytes(64))
		require.NoError(t, err)
	}

	batches, err := f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 10, len(batches))

	// Should create 11th batch Batch 10 and not prune anything.
	f.AddBlock()
	err = f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(t, 1)[0])
	require.NoError(t, err)
	err = f.batchingKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 11, len(batches))

	// Should create 12th batch Batch 11 and prune Batch 0.
	f.AddBlock()
	err = f.batchingKeeper.SetDataResultForBatching(f.Context(), generateDataResults(t, 1)[0])
	require.NoError(t, err)
	err = f.batchingKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	batches, err = f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 11, len(batches))
	require.Equal(t, uint64(1), batches[0].BatchNumber)
	f.checkNoBatchData(t, 0)
	for i := 1; i <= 9; i++ {
		f.checkBatchData(t, uint64(i), true)
	}
	f.checkBatchData(t, 10, false) // did not add signatures for latest two batches
	f.checkBatchData(t, 11, false)
}

func TestPruningMockedUpgrade(t *testing.T) {
	f := initFixture(t)

	f.addBatchSigningValidators(t, 10)

	err := f.pubKeyKeeper.SetProvingScheme(f.Context(), pubkeytypes.ProvingScheme{
		Index:       uint32(sedatypes.SEDAKeyIndexSecp256k1),
		IsActivated: true,
	})
	require.NoError(t, err)

	// Adjust the global variable for the test.
	original := keeper.NumBatchesToKeep
	defer func() {
		keeper.NumBatchesToKeep = original
	}()
	keeper.NumBatchesToKeep = 10

	err = f.batchingKeeper.SetParams(f.Context(), types.Params{
		MaxBatchPrunePerBlock:            15,
		MaxLegacyDataResultPrunePerBlock: 80,
	})
	require.NoError(t, err)

	// Create 30 batches with 10 data results each before mock upgrade.
	// We simulate the chain before the upgrade by using legacy functions.
	for range 30 {
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
		err = f.batchingKeeper.SetBatchSigSecp256k1(f.Context(), batch.BatchNumber, valEntries[0].ValidatorAddress, generateRandomBytes(64))
		require.NoError(t, err)
	}

	// Mock upgrade at height 30:
	// - Upgrade handler should set batchNumberAtUpgrade and hasPruningCaughtUp.
	err = f.batchingKeeper.SetBatchNumberAtUpgrade(f.Context())
	require.NoError(t, err)
	err = f.batchingKeeper.SetHasPruningCaughtUp(f.Context(), false)
	require.NoError(t, err)

	// Block 31:
	// - Creates 31st batch Batch 30
	// - Basic pruning prunes Batch 20
	// - Batch prunes Batches 0-14
	f.BatchingEndBlock(t, 10)

	batches, err := f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, 15, len(batches))
	require.Equal(t, uint64(15), batches[0].BatchNumber)
	require.Equal(t, uint64(30), batches[len(batches)-1].BatchNumber)
	for i := 0; i <= 30; i++ {
		if i <= 14 || i == 20 {
			f.checkNoBatchData(t, uint64(i))
		} else {
			f.checkBatchData(t, uint64(i), false)
		}
	}
	f.checkNumLegacyDataResults(t, 300)

	hasCaughtUp, err := f.batchingKeeper.HasPruningCaughtUp(f.Context())
	require.NoError(t, err)
	require.False(t, hasCaughtUp)

	// Block 32~39:
	//   - Block 32:
	//     - Creates 32nd batch Batch 31
	//     - Basic pruning prunes Batch 21
	//     - Batch pruning prunes Batches 15-21 (20 & 21 already pruned by basic pruning)
	//     - Legacy data result pruning starts next block
	for i := range 8 {
		f.BatchingEndBlock(t, 10)

		batches, err = f.batchingKeeper.GetAllBatches(f.Context())
		require.NoError(t, err)
		require.Equal(t, 10, len(batches))
		require.Equal(t, uint64(22+i), batches[0].BatchNumber)
		require.Equal(t, uint64(31+i), batches[len(batches)-1].BatchNumber)
		for j := 0; j <= 21+i; j++ {
			f.checkNoBatchData(t, uint64(j))
		}
		for j := 22 + i; j <= 31+i; j++ {
			f.checkBatchData(t, uint64(j), false)
		}
		f.checkNumLegacyDataResults(t, max(300-80*i, 0))

		hasCaughtUp, err = f.batchingKeeper.HasPruningCaughtUp(f.Context())
		require.NoError(t, err)
		require.True(t, hasCaughtUp)
	}

	// Block 40 - 43:
	// - Batch creation at every block but number of batches stays at 10 with basic pruning.
	for i := range 4 {
		f.BatchingEndBlock(t, 10)

		batches, err = f.batchingKeeper.GetAllBatches(f.Context())
		require.NoError(t, err)
		require.Equal(t, 10, len(batches))
		require.Equal(t, uint64(30+i), batches[0].BatchNumber)
		require.Equal(t, uint64(39+i), batches[len(batches)-1].BatchNumber)
		for j := 0; j <= 29+i; j++ {
			f.checkNoBatchData(t, uint64(j))
		}
		for j := 30 + i; j <= 39+i; j++ {
			f.checkBatchData(t, uint64(j), false)
		}
		f.checkNumLegacyDataResults(t, 0)

		hasCaughtUp, err = f.batchingKeeper.HasPruningCaughtUp(f.Context())
		require.NoError(t, err)
		require.True(t, hasCaughtUp)
	}

	// Block 44 without batch creation
	f.BatchingEndBlock(t, 0)
	f.checkNumLegacyDataResults(t, 0)

	// Block 45 without batch creation
	f.BatchingEndBlock(t, 5)
	f.checkNumLegacyDataResults(t, 0)

}

// BatchingEndBlock adds a given number of data results to the store and executes
// batching EndBlock. Note if there is no change in data result or validator tree
// root, no new batch is created.
func (f *fixture) BatchingEndBlock(t *testing.T, numDataResults int) {
	f.AddBlock()
	if numDataResults > 0 {
		dataResults := generateDataResults(t, numDataResults)
		for _, dataResult := range dataResults {
			err := f.batchingKeeper.SetDataResultForBatching(f.Context(), dataResult)
			require.NoError(t, err)
		}
	}
	err := f.batchingKeeper.EndBlock(f.Context())
	require.NoError(t, err)
}

func (f *fixture) checkNoBatchData(t *testing.T, batchNum uint64) {
	batch, err := f.batchingKeeper.GetBatchByBatchNumber(f.Context(), batchNum)
	require.ErrorIs(t, err, collections.ErrNotFound, "batchNum %d", batchNum)
	dataEntries, err := f.batchingKeeper.GetDataResultTreeEntries(f.Context(), batchNum)
	require.ErrorIs(t, err, collections.ErrNotFound, "batchNum %d", batchNum)
	valEntries, _ := f.batchingKeeper.GetValidatorTreeEntries(f.Context(), batchNum)
	// require.ErrorIs(t, err, collections.ErrNotFound) // this function does not error even if there are no entries.
	sigs, _ := f.batchingKeeper.GetBatchSignatures(f.Context(), batchNum)
	// require.ErrorIs(t, err, collections.ErrNotFound) // this function does not error even if there are no entries.

	// TODO batchDataResults and dataResults

	require.Empty(t, batch, "batchNum: %d", batchNum)
	require.Empty(t, dataEntries, "batchNum: %d", batchNum)
	require.Empty(t, valEntries, "batchNum: %d", batchNum)
	require.Empty(t, sigs, "batchNum: %d", batchNum)
}

func (f *fixture) checkBatchData(t *testing.T, batchNum uint64, checkSigs bool) {
	batch, err := f.batchingKeeper.GetBatchByBatchNumber(f.Context(), batchNum)
	require.NoError(t, err)
	dataEntries, err := f.batchingKeeper.GetDataResultTreeEntries(f.Context(), batchNum)
	require.NoError(t, err)
	valEntries, err := f.batchingKeeper.GetValidatorTreeEntries(f.Context(), batchNum)
	require.NoError(t, err)

	// TODO batchDataResults and dataResults

	require.NotEmpty(t, batch, "batch number %d batch should not be empty", batchNum)
	require.NotEmpty(t, dataEntries, "batch number %d data entriesshould not be empty", batchNum)
	require.NotEmpty(t, valEntries, "batch number %d validator entries should not be empty", batchNum)
	if checkSigs {
		sigs, err := f.batchingKeeper.GetBatchSignatures(f.Context(), batchNum)
		require.NoError(t, err)
		require.NotEmpty(t, sigs, "batch number %d signatures should not be empty", batchNum)
	}
}

func (f *fixture) checkNumLegacyDataResults(t *testing.T, expectedNum int) {
	dataResults, err := f.batchingKeeper.GetLegacyDataResults(f.Context(), true)
	require.NoError(t, err)
	require.Equal(t, expectedNum, len(dataResults))
}
