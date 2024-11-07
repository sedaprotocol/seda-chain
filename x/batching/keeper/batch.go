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
func (k Keeper) SetNewBatch(ctx context.Context, batch types.Batch, dataEntries [][]byte, valEntries []ValidatorTreeEntry) error {
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

	err = k.setTreeEntriesForBatch(ctx, newBatchNum, dataEntries, valEntries)
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

type ValidatorTreeEntry struct {
	ValAddress sdk.ValAddress
	Entry      []byte
}

// setTreeEntry stores a tree entry.
func (k Keeper) setTreeEntry(ctx context.Context, entry types.TreeEntry) error {
	return k.treeEntries.Set(ctx, collections.Join(entry.K1, entry.K2), entry.Entry)
}

// setTreeEntriesForBatch stores the data result entries and validator
// entries for a given batch.
func (k Keeper) setTreeEntriesForBatch(ctx context.Context, batchNum uint64, dataEntries [][]byte, valEntries []ValidatorTreeEntry) error {
	for _, entry := range dataEntries {
		err := k.treeEntries.Set(ctx, collections.Join(batchNum, []byte{}), entry)
		if err != nil {
			return err
		}
	}
	for _, entry := range valEntries {
		err := k.treeEntries.Set(ctx, collections.Join(batchNum, entry.ValAddress.Bytes()), entry.Entry)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetValidatorTreeEntry returns the tree entry of a given validator
// for a specified batch
func (k Keeper) GetValidatorTreeEntry(ctx context.Context, batchNum uint64, valAddress sdk.ValAddress) ([]byte, error) {
	return k.treeEntries.Get(ctx, collections.Join(batchNum, valAddress.Bytes()))
}

// GetTreeEntriesForBatch returns the tree entries corresponding to
// the givenbatch number.
func (k Keeper) GetTreeEntriesForBatch(ctx context.Context, batchNum uint64) (types.TreeEntries, error) {
	rng := collections.NewPrefixedPairRange[uint64, []byte](batchNum)
	itr, err := k.treeEntries.Iterate(ctx, rng)
	if err != nil {
		return types.TreeEntries{}, err
	}
	defer itr.Close()

	kvs, err := itr.KeyValues()
	if err != nil {
		return types.TreeEntries{}, err
	}
	if len(kvs) == 0 {
		return types.TreeEntries{}, collections.ErrNotFound
	}

	var dataEntries, valEntries [][]byte
	for _, kv := range kvs {
		if len(kv.Key.K2()) == 0 {
			dataEntries = append(dataEntries, kv.Value)
		} else {
			valEntries = append(valEntries, kv.Value)
		}
	}
	return types.TreeEntries{
		BatchNumber:       batchNum,
		DataResultEntries: dataEntries,
		ValidatorEntries:  valEntries,
	}, nil
}

// IterateBatches iterates over the tree entries and performs a given
// callback function.
func (k Keeper) IterateTreeEntries(ctx sdk.Context, callback func(types.TreeEntry) (stop bool)) error {
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

		if callback(types.TreeEntry{
			K1:    kv.Key.K1(),
			K2:    kv.Key.K2(),
			Entry: kv.Value,
		}) {
			break
		}
	}
	return nil
}

// GetAllTreeEntries retrieves all tree entries from the store.
func (k Keeper) GetAllTreeEntries(ctx sdk.Context) ([]types.TreeEntry, error) {
	var entries []types.TreeEntry
	err := k.IterateTreeEntries(ctx, func(entry types.TreeEntry) bool {
		entries = append(entries, entry)
		return false
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// SetBatchSignatures stores a validator's signatures of a batch.
func (k Keeper) SetBatchSignatures(ctx context.Context, sigs types.BatchSignatures) error {
	valAddr, err := k.validatorAddressCodec.StringToBytes(sigs.ValidatorAddr)
	if err != nil {
		return err
	}
	return k.batchSignatures.Set(ctx, collections.Join(sigs.BatchNumber, valAddr), sigs)
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

// GetAllBatchSignatures returns all batch signatures in the store.
func (k Keeper) GetAllBatchSignatures(ctx context.Context) ([]types.BatchSignatures, error) {
	itr, err := k.batchSignatures.Iterate(ctx, nil)
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
