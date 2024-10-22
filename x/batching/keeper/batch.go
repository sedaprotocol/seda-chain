package keeper

import (
	"context"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (k Keeper) setCurrentBatchNum(ctx context.Context, batchNum uint64) error {
	return k.currentBatchNumber.Set(ctx, batchNum)
}

// incrementCurrentBatchNum increments the batch number sequence and
// returns the new number.
func (k Keeper) incrementCurrentBatchNum(ctx context.Context) (uint64, error) {
	next, err := k.currentBatchNumber.Next(ctx)
	return next + 1, err
}

func (k Keeper) GetCurrentBatchNum(ctx context.Context) (uint64, error) {
	batchNum, err := k.currentBatchNumber.Peek(ctx)
	if err != nil {
		return 0, err
	}
	return batchNum, nil
}

func (k Keeper) setBatch(ctx context.Context, batch types.Batch) error {
	return k.batches.Set(ctx, batch.BlockHeight, batch)
}

// SetNewBatch increments the current batch number and stores a given
// batch at that index. It also stores the given data result tree entries
// and validator tree entries. It returns an error if a batch already
// exists at the given batch's block height or if the given batch's
// batch number does not match the next batch number.
func (k Keeper) SetNewBatch(ctx context.Context, batch types.Batch, dataEntries, valEntries [][]byte) error {
	found, err := k.batches.Has(ctx, batch.BlockHeight)
	if err != nil {
		return err
	}
	if found {
		return types.ErrBatchAlreadyExists.Wrapf("batch block height %d", batch.BlockHeight)
	}

	newBatchNum, err := k.incrementCurrentBatchNum(ctx)
	if err != nil {
		return err
	}
	if batch.BatchNumber != newBatchNum {
		return types.ErrInvalidBatchNumber.Wrapf("got %d; expected %d", batch.BatchNumber, newBatchNum)
	}
	batch.BatchNumber = newBatchNum

	err = k.setTreeEntries(ctx, newBatchNum, dataEntries, valEntries)
	if err != nil {
		return err
	}
	return k.batches.Set(ctx, batch.BlockHeight, batch)
}

func (k Keeper) GetBatchForHeight(ctx context.Context, blockHeight int64) (types.Batch, error) {
	batch, err := k.batches.Get(ctx, blockHeight)
	if err != nil {
		return types.Batch{}, err
	}
	return batch, nil
}

// GetLatestBatch returns the most recently created batch. If batching
// has not begun, it returns an error ErrBatchingHasNotStarted.
func (k Keeper) GetLatestBatch(ctx context.Context) (types.Batch, error) {
	currentBatchNum, err := k.currentBatchNumber.Peek(ctx)
	if err != nil {
		return types.Batch{}, err
	}
	if currentBatchNum == collections.DefaultSequenceStart {
		return types.Batch{}, types.ErrBatchingHasNotStarted
	}
	return k.GetBatchByBatchNumber(ctx, currentBatchNum)
}

func (k Keeper) GetBatchByBatchNumber(ctx context.Context, batchNumber uint64) (types.Batch, error) {
	blockHeight, err := k.batches.Indexes.Number.MatchExact(ctx, batchNumber)
	if err != nil {
		return types.Batch{}, err
	}
	return k.batches.Get(ctx, blockHeight)
}

// IterateBatches iterates over the batches and performs a given
// callback function.
func (k Keeper) IterateBatches(ctx sdk.Context, callback func(types.Batch) (stop bool)) error {
	iter, err := k.batches.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if callback(kv.Value) {
			break
		}
	}
	return nil
}

// GetAllBatches returns all batches in the store.
func (k Keeper) GetAllBatches(ctx sdk.Context) ([]types.Batch, error) {
	var batches []types.Batch
	err := k.IterateBatches(ctx, func(batch types.Batch) bool {
		batches = append(batches, batch)
		return false
	})
	if err != nil {
		return nil, err
	}
	return batches, nil
}

// setTreeEntries stores the data result entries and validator entries
// using the given batch number as the key.
func (k Keeper) setTreeEntries(ctx context.Context, batchNum uint64, dataEntries, valEntries [][]byte) error {
	return k.treeEntries.Set(
		ctx,
		batchNum,
		types.TreeEntries{
			BatchNumber:       batchNum,
			DataResultEntries: dataEntries,
			ValidatorEntries:  valEntries,
		},
	)
}

// GetTreeEntries returns the tree entries corresponding to the given
// batch number.
func (k Keeper) GetTreeEntries(ctx context.Context, batchNum uint64) (types.TreeEntries, error) {
	entries, err := k.treeEntries.Get(ctx, batchNum)
	if err != nil {
		return types.TreeEntries{}, err
	}
	return entries, nil
}

// IterateBatches iterates over the tree entries and performs a given
// callback function.
func (k Keeper) IterateTreeEntries(ctx sdk.Context, callback func(types.TreeEntries) (stop bool)) error {
	iter, err := k.treeEntries.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if callback(kv.Value) {
			break
		}
	}
	return nil
}

// GetAllTreeEntries retrieves all tree entries from the store.
func (k Keeper) GetAllTreeEntries(ctx sdk.Context) ([]types.TreeEntries, error) {
	var entries []types.TreeEntries
	err := k.IterateTreeEntries(ctx, func(entry types.TreeEntries) bool {
		entries = append(entries, entry)
		return false
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// SetBatchSignatures stores a validator's signatures of a batch.
func (k Keeper) SetBatchSignatures(ctx context.Context, batchNum uint64, sigs types.BatchSignatures) error {
	valAddr, err := k.validatorAddressCodec.StringToBytes(sigs.ValidatorAddr)
	if err != nil {
		return err
	}
	return k.batchSignatures.Set(ctx, collections.Join(batchNum, valAddr), sigs)
}

// GetBatchSignatures retrieves the batch signatures by a given
// validator at a given batch number.
func (k Keeper) GetBatchSignatures(ctx context.Context, batchNum uint64, validatorAddr string) (types.BatchSignatures, error) {
	valAddr, err := k.validatorAddressCodec.StringToBytes(validatorAddr)
	if err != nil {
		return types.BatchSignatures{}, err
	}
	sigs, err := k.batchSignatures.Get(ctx, collections.Join(batchNum, valAddr))
	if err != nil {
		return types.BatchSignatures{}, err
	}
	return sigs, err
}

// GetBatchSigsForBatch returns all signatures of a given batch.
func (k Keeper) GetBatchSigsForBatch(ctx context.Context, batchNum uint64) ([]types.BatchSignatures, error) {
	rng := collections.NewPrefixedPairRange[uint64, []byte](batchNum)
	itr, err := k.batchSignatures.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	kvs, err := itr.KeyValues()
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, collections.ErrNotFound
	}

	sigs := make([]types.BatchSignatures, len(kvs))
	for i, kv := range kvs {
		sigs[i] = kv.Value
	}
	return sigs, err
}
