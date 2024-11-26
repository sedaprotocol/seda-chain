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
func (k Keeper) SetNewBatch(ctx context.Context, batch types.Batch, dataEntries types.DataResultTreeEntries, valEntries []types.ValidatorTreeEntry) error {
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

	err = k.setDataResultTreeEntry(ctx, newBatchNum, dataEntries)
	if err != nil {
		return err
	}

	for _, valEntry := range valEntries {
		err = k.setValidatorTreeEntry(ctx, newBatchNum, valEntry)
		if err != nil {
			return err
		}
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

func (k Keeper) GetBatchByBatchNumber(ctx context.Context, batchNumber uint64) (types.Batch, error) {
	blockHeight, err := k.batches.Indexes.Number.MatchExact(ctx, batchNumber)
	if err != nil {
		return types.Batch{}, err
	}
	return k.batches.Get(ctx, blockHeight)
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

func (k Keeper) setValidatorTreeEntry(ctx context.Context, batchNum uint64, entry types.ValidatorTreeEntry) error {
	err := k.validatorTreeEntries.Set(ctx, collections.Join(batchNum, entry.ValidatorAddress.Bytes()), entry)
	if err != nil {
		return err
	}
	return nil
}

// GetValidatorTreeEntry returns the tree entry of a given validator
// for a specified batch
func (k Keeper) GetValidatorTreeEntry(ctx context.Context, batchNum uint64, valAddress sdk.ValAddress) (types.ValidatorTreeEntry, error) {
	valEntry, err := k.validatorTreeEntries.Get(ctx, collections.Join(batchNum, valAddress.Bytes()))
	if err != nil {
		return types.ValidatorTreeEntry{}, err
	}
	return valEntry, nil
}

func (k Keeper) setDataResultTreeEntry(ctx context.Context, batchNum uint64, dataEntries types.DataResultTreeEntries) error {
	err := k.dataResultTreeEntries.Set(ctx, batchNum, dataEntries)
	if err != nil {
		return err
	}
	return nil
}

// GetTreeEntriesForBatch returns the tree entries of a given batch.
func (k Keeper) GetTreeEntriesForBatch(ctx context.Context, batchNum uint64) (types.BatchTreeEntries, error) {
	// Get data result tree entries.
	dataEntries, err := k.dataResultTreeEntries.Get(ctx, batchNum)
	if err != nil {
		return types.BatchTreeEntries{}, err
	}

	// Get validator tree entries.
	valRng := collections.NewPrefixedPairRange[uint64, []byte](batchNum)
	valItr, err := k.validatorTreeEntries.Iterate(ctx, valRng)
	if err != nil {
		return types.BatchTreeEntries{}, err
	}
	defer valItr.Close()

	valKvs, err := valItr.KeyValues()
	if err != nil {
		return types.BatchTreeEntries{}, err
	}
	var valEntries []types.ValidatorTreeEntry
	for _, kv := range valKvs {
		valEntries = append(valEntries, kv.Value)
	}

	return types.BatchTreeEntries{
		BatchNumber:       batchNum,
		DataResultEntries: dataEntries,
		ValidatorEntries:  valEntries,
	}, nil
}

// SetBatchSigSecp256k1 stores a validator's secp256k1 signatures of
// a batch.
func (k Keeper) SetBatchSigSecp256k1(ctx context.Context, batchNum uint64, operatorAddress string, signature []byte) error {
	valAddr, err := k.validatorAddressCodec.StringToBytes(operatorAddress)
	if err != nil {
		return err
	}
	entry, err := k.GetValidatorTreeEntry(ctx, batchNum, valAddr)
	if err != nil {
		return err
	}
	entry.Secp256K1.Signature = signature
	return k.setValidatorTreeEntry(ctx, batchNum, entry)
}
