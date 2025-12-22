package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBatchNumberAtUpgrade sets the latest batch number at the time of the upgrade.
func (k Keeper) SetBatchNumberAtUpgrade(ctx sdk.Context) error {
	// Latest batch number is the current batch number minus 1 because
	// the current batch number has not been used yet.
	currentBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return err
	}
	return k.batchNumberAtUpgrade.Set(ctx, currentBatchNum-1)
}

func (k Keeper) GetBatchNumberAtUpgrade(ctx sdk.Context) (uint64, error) {
	return k.batchNumberAtUpgrade.Get(ctx)
}

func (k Keeper) SetHasPruningCaughtUp(ctx sdk.Context, hasCaughtUp bool) error {
	return k.hasPruningCaughtUp.Set(ctx, hasCaughtUp)
}

func (k Keeper) HasPruningCaughtUp(ctx sdk.Context) (bool, error) {
	return k.hasPruningCaughtUp.Get(ctx)
}

// TryPruneBatch attempts to prune the given batch and all of its associated data.
func (k Keeper) TryPruneBatch(ctx sdk.Context, batchNum uint64) error {
	batch, err := k.GetBatchByBatchNumber(ctx, batchNum)
	if err != nil {
		return err
	}
	batchHeight := batch.BlockHeight

	err = k.batches.Remove(ctx, batchHeight)
	if err != nil {
		return err
	}
	err = k.dataResultTreeEntries.Remove(ctx, batchNum)
	if err != nil {
		return err
	}

	valRng := new(collections.Range[collections.Pair[uint64, []byte]]).Prefix(collections.PairPrefix[uint64, []byte](batchNum))
	err = k.validatorTreeEntries.Clear(ctx, valRng)
	if err != nil {
		return err
	}
	err = k.batchSignatures.Clear(ctx, valRng)
	if err != nil {
		return err
	}

	dataResults, err := k.GetBatchDataResults(ctx, batchNum)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			k.Logger(ctx).Info("cannot prune batch because schema change has not been applied", "batch_num", batchNum)
			return nil
		}
		return err
	}
	for _, item := range dataResults.DataRequestIdHeights {
		err = k.RemoveBatchAssignment(ctx, item.DataRequestId, item.DataRequestHeight)
		if err != nil {
			return err
		}
		err = k.RemoveDataResult(ctx, true, item.DataRequestId, item.DataRequestHeight)
		if err != nil {
			return err
		}
	}
	err = k.RemoveBatchDataResults(ctx, batchNum)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("single pruned batch", "batch_num", batchNum)
	return nil
}

// BatchPruneBatches prunes batches and their associated data, except for data
// results, in batches based on the module parameters NumBatchesToKeep and
// MaxBatchPrunePerBlock.
// It returns the batch number of the last batch that has been confirmed to have
// been pruned.
func (k Keeper) BatchPruneBatches(ctx sdk.Context, numBatchesToKeep, maxBatchPrunePerBlock, batchNumAtUpgrade uint64) (uint64, error) {
	if maxBatchPrunePerBlock == 0 {
		k.Logger(ctx).Info("skip batch pruning", "max_batch_prune_per_block", maxBatchPrunePerBlock)
		return 0, nil
	}

	// Prune up to, but not including, current batch number minus numBatchesToKeep.
	currentBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return 0, err
	}
	if currentBatchNum <= numBatchesToKeep {
		k.Logger(ctx).Info("skip batch pruning", "current_batch_num", currentBatchNum, "num_batches_to_keep", numBatchesToKeep)
		return 0, nil
	}

	rngEnd := min(currentBatchNum-numBatchesToKeep, batchNumAtUpgrade+1)
	rng := new(collections.Range[uint64]).EndExclusive(rngEnd)
	iter, err := k.batches.Indexes.Number.Iterate(ctx, rng)
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	var firstKey *collections.Pair[uint64, int64]
	var pruneCount uint64
	var lastPrunedBatchNum uint64
	for ; iter.Valid(); iter.Next() {
		fullKey, err := iter.FullKey()
		if err != nil {
			return 0, err
		}
		if firstKey == nil {
			firstKey = &fullKey
		}

		batchNum, batchHeight := fullKey.K1(), fullKey.K2()
		if batchNum >= currentBatchNum-numBatchesToKeep {
			// Should not happen given the range configuration.
			break
		}

		err = k.batches.Remove(ctx, batchHeight)
		if err != nil {
			return 0, err
		}
		k.Logger(ctx).Info("pruned batch", "batch_num", batchNum)

		lastPrunedBatchNum = batchNum

		pruneCount++
		if pruneCount == maxBatchPrunePerBlock {
			break
		}
	}

	if firstKey == nil {
		k.Logger(ctx).Info("no batches to prune")
		// This means all batches up to batch number rngEnd - 1 have been pruned.
		// Note we subtract 1 because rngEnd is exclusive.
		return rngEnd - 1, nil
	}

	dataRng := new(collections.Range[uint64]).EndExclusive(firstKey.K1() + pruneCount)
	err = k.dataResultTreeEntries.Clear(ctx, dataRng)
	if err != nil {
		return 0, err
	}

	valRng := new(collections.Range[collections.Pair[uint64, []byte]]).
		EndExclusive(collections.PairPrefix[uint64, []byte](firstKey.K1() + pruneCount))
	err = k.validatorTreeEntries.Clear(ctx, valRng)
	if err != nil {
		return 0, err
	}
	err = k.batchSignatures.Clear(ctx, valRng)
	if err != nil {
		return 0, err
	}

	return lastPrunedBatchNum, nil
}

func (k Keeper) PruneLegacyDataResults(ctx sdk.Context, maxDataResultPrunePerBlock uint64) error {
	if maxDataResultPrunePerBlock == 0 {
		k.Logger(ctx).Info("skip legacy data result pruning", "max_data_results_to_check_for_prune", maxDataResultPrunePerBlock)
		return nil
	}

	iter, err := k.legacyDataResults.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	var numPruned uint64
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		err = k.RemoveLegacyDataResult(ctx, true, kv.Value.DrId, kv.Value.DrBlockHeight)
		if err != nil {
			return err
		}
		err = k.RemoveBatchAssignment(ctx, kv.Value.DrId, kv.Value.DrBlockHeight)
		if err != nil {
			return err
		}

		numPruned++
		if numPruned == maxDataResultPrunePerBlock {
			break
		}
	}

	k.Logger(ctx).Info("pruned legacy data results", "num_pruned", numPruned)
	return nil
}
