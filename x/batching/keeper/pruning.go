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
	err = k.batchNumberAtUpgrade.Set(ctx, currentBatchNum-1)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info("set batch number at upgrade", "batch_number", currentBatchNum-1)
	return nil
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

// BasicPruneBatch prunes a given batch and its associated tree entries and
// signatures. It also prunes associated data results if the batchDataResults
// mapping is available. It returns without error if the batch does not exist
// (must have been pruned by the batch pruning strategy).
func (k Keeper) BasicPruneBatch(ctx sdk.Context, batchNumber uint64) error {
	k.Logger(ctx).Info("[basic pruning strategy] pruning a batch", "batch_num", batchNumber)

	batch, err := k.GetBatchByBatchNumber(ctx, batchNumber)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Batch may have been pruned already.
			k.Logger(ctx).Info("[basic pruning strategy] batch not found", "batch_num", batchNumber)
			return nil
		}
		return err
	}
	batchHeight := batch.BlockHeight

	err = k.batches.Remove(ctx, batchHeight)
	if err != nil {
		return err
	}
	err = k.dataResultTreeEntries.Remove(ctx, batchNumber)
	if err != nil {
		return err
	}

	valRng := new(collections.Range[collections.Pair[uint64, []byte]]).Prefix(collections.PairPrefix[uint64, []byte](batchNumber))
	err = k.validatorTreeEntries.Clear(ctx, valRng)
	if err != nil {
		return err
	}
	err = k.batchSignatures.Clear(ctx, valRng)
	if err != nil {
		return err
	}

	dataResults, err := k.GetBatchDataResults(ctx, batchNumber)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Batches created before the upgrade do not have batchDataResults
			// mapping, so we resort to PruneLegacyDataResults().
			k.Logger(ctx).Info("[basic pruning strategy] skip pruning data results", "batch_num", batchNumber)
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
	err = k.RemoveBatchDataResults(ctx, batchNumber)
	if err != nil {
		return err
	}

	return nil
}

// BatchPruneBatches prunes up to MaxBatchPrunePerBlock batches and their associated
// data except for data results.
// It returns a boolean hasCaughtUp if either of the following conditions is met:
// (i) All batches up to batchNumberAtUpgrade have been pruned.
// (ii) All batches up to (currentBatchNum - numBatchesToKeep) have been pruned.
func (k Keeper) BatchPruneBatches(ctx sdk.Context, numBatchesToKeep, maxBatchPrunePerBlock uint64) (bool, error) {
	if maxBatchPrunePerBlock == 0 {
		k.Logger(ctx).Info("[batch pruning strategy] disabled")
		return false, nil
	}

	batchNumAtUpgrade, err := k.GetBatchNumberAtUpgrade(ctx)
	if err != nil {
		return false, nil
	}
	if batchNumAtUpgrade == 0 {
		k.Logger(ctx).Info("[batch pruning strategy] skipped due to lack of upgrade")
		return false, nil
	}

	currentBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return false, err
	}
	if currentBatchNum <= numBatchesToKeep {
		k.Logger(ctx).Info("[batch pruning strategy] skipped", "current_batch_num", currentBatchNum, "num_batches_to_keep", numBatchesToKeep)
		return false, nil
	}

	rngEnd := min(currentBatchNum-numBatchesToKeep, batchNumAtUpgrade+1)
	rng := new(collections.Range[uint64]).EndExclusive(rngEnd)
	iter, err := k.batches.Indexes.Number.Iterate(ctx, rng)
	if err != nil {
		return false, err
	}
	defer iter.Close()

	var pruneCount, lastPrunedBatchNum uint64
	for ; iter.Valid(); iter.Next() {
		fullKey, err := iter.FullKey()
		if err != nil {
			return false, err
		}

		batchNum, batchHeight := fullKey.K1(), fullKey.K2()
		if batchNum >= currentBatchNum-numBatchesToKeep {
			// Should not happen given the range configuration.
			break
		}

		err = k.batches.Remove(ctx, batchHeight)
		if err != nil {
			return false, err
		}
		k.Logger(ctx).Debug("[batch pruning strategy] pruned a batch", "batch_num", batchNum)

		lastPrunedBatchNum = batchNum

		pruneCount++
		if pruneCount == maxBatchPrunePerBlock {
			break
		}
	}

	if pruneCount == 0 {
		k.Logger(ctx).Info("[batch pruning strategy] no batches to prune")
		return true, nil
	}

	dataRng := new(collections.Range[uint64]).EndExclusive(lastPrunedBatchNum + 1)
	err = k.dataResultTreeEntries.Clear(ctx, dataRng)
	if err != nil {
		return false, err
	}

	valRng := new(collections.Range[collections.Pair[uint64, []byte]]).
		EndExclusive(collections.PairPrefix[uint64, []byte](lastPrunedBatchNum + 1))
	err = k.validatorTreeEntries.Clear(ctx, valRng)
	if err != nil {
		return false, err
	}
	err = k.batchSignatures.Clear(ctx, valRng)
	if err != nil {
		return false, err
	}

	k.Logger(ctx).Info("[batch pruning strategy] pruned batches", "count", pruneCount)
	return pruneCount < maxBatchPrunePerBlock, nil
}

func (k Keeper) PruneLegacyDataResults(ctx sdk.Context, maxDataResultPrunePerBlock uint64) error {
	if maxDataResultPrunePerBlock == 0 {
		k.Logger(ctx).Info("[legacy data result pruning] disabled")
		return nil
	}

	iter, err := k.legacyDataResults.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	var pruneCount uint64
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

		pruneCount++
		if pruneCount == maxDataResultPrunePerBlock {
			break
		}
	}

	k.Logger(ctx).Info("[legacy data result pruning] pruned legacy data results", "count", pruneCount)
	return nil
}
