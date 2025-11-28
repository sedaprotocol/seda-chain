package keeper

import (
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/sha3"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PruneBatch attempts to prune a given batch and all its associated data. This
// function requires the map batchDataResults, which was added in a chain upgrade.
func (k Keeper) PruneBatch(ctx sdk.Context, batchNum uint64) error {
	batch, err := k.GetBatchByBatchNumber(ctx, batchNum)
	if err != nil {
		// The given batch may have been pruned already.
		if errors.Is(err, collections.ErrNotFound) {
			k.Logger(ctx).Info("cannot prune batch because it does not exist", "batch_num", batchNum)
			return nil
		}
		return err
	}
	batchHeight := batch.BlockHeight

	// This functino requires the map batchDataResults, which was added in a chain
	// upgrade. If the mapping does not exist for the given batch, we resort to
	// BatchPruneBatches() to prune it.
	found, err := k.batchDataResults.Has(ctx, batchNum)
	if err != nil {
		return err
	}
	if !found {
		k.Logger(ctx).Info("cannot prune batch because schema change has not been applied", "batch_num", batchNum)
		return nil
	}

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

	dataResults, err := k.batchDataResults.Get(ctx, batchNum)
	if err != nil {
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

// BatchPruneBatches prunes batches and their associated data in batches based on
// the module parameters NumBatchesToKeep and MaxBatchPrunePerBlock. It returns the
// batch number of the last pruned batch.
func (k Keeper) BatchPruneBatches(ctx sdk.Context, numBatchesToKeep, maxBatchPrunePerBlock uint64) (uint64, error) {
	if maxBatchPrunePerBlock == 0 {
		k.Logger(ctx).Info("skip batch pruning", "max_batch_prune_per_block", maxBatchPrunePerBlock)
		return 0, nil
	}

	currentBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return 0, err
	}
	if currentBatchNum <= numBatchesToKeep {
		k.Logger(ctx).Info("skip batch pruning", "current_batch_num", currentBatchNum, "num_batches_to_keep", numBatchesToKeep)
		return 0, nil
	}

	rng := new(collections.Range[uint64]).EndExclusive(currentBatchNum - numBatchesToKeep)
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
			// Should not happen because of the range configuration.
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
		// This means nothing was pruned.
		k.Logger(ctx).Info("no batches to prune")
		return 0, nil
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

func (k Keeper) BatchPruneDataResults(ctx sdk.Context, maxDataResultsToCheckForPrune, lastRemovedBatchNum uint64) error {
	if maxDataResultsToCheckForPrune == 0 || lastRemovedBatchNum == 0 {
		k.Logger(ctx).Info("skip data result pruning", "max_data_results_to_check_for_prune", maxDataResultsToCheckForPrune, "last_removed_batch_num", lastRemovedBatchNum)
		return nil
	}

	// Use hash of last commit hash as starting point of the range.
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(ctx.BlockHeader().LastCommitHash)
	hash := hasher.Sum(nil)

	var rng *collections.Range[collections.Triple[bool, string, uint64]]
	if ctx.BlockHeight()%2 == 0 {
		rng = new(collections.Range[collections.Triple[bool, string, uint64]]).
			StartInclusive(collections.TripleSuperPrefix[bool, string, uint64](true, hex.EncodeToString(hash)))
	} else {
		rng = new(collections.Range[collections.Triple[bool, string, uint64]]).
			EndInclusive(collections.TripleSuperPrefix[bool, string, uint64](true, hex.EncodeToString(hash))).
			Descending()
	}

	iter, err := k.dataResults.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()

	var numChecked, numPruned uint64
	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		batchNum, err := k.GetBatchAssignment(ctx, kv.Value.DrId, kv.Value.DrBlockHeight)
		if err != nil {
			return err
		}

		if batchNum <= lastRemovedBatchNum {
			err = k.RemoveDataResult(ctx, true, kv.Value.DrId, kv.Value.DrBlockHeight)
			if err != nil {
				return err
			}
			err = k.batchAssignments.Remove(ctx, collections.Join(kv.Value.DrId, kv.Value.DrBlockHeight))
			if err != nil {
				return err
			}
			numPruned++
		}

		numChecked++
		if numChecked == maxDataResultsToCheckForPrune {
			break
		}
	}

	k.Logger(ctx).Info("pruned data results", "num_checked", numChecked, "num_pruned", numPruned)
	return nil
}
