package keeper

import (
	"context"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/abci"
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

// SetNewBatch stores a new batch and its associated data at current batch number
// and increments the current batch number. If successful, it returns the batch number
// of the newly created batch.
func (k Keeper) SetNewBatch(ctx sdk.Context, batch types.Batch, dataEntries types.DataResultTreeEntries, valEntries []types.ValidatorTreeEntry) (uint64, error) {
	found, err := k.batches.Has(ctx, batch.BlockHeight)
	if err != nil {
		return 0, err
	}
	if found {
		return 0, types.ErrBatchAlreadyExists.Wrapf("batch block height %d", batch.BlockHeight)
	}

	batchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return 0, err
	}
	if batch.BatchNumber != batchNum {
		return 0, types.ErrInvalidBatchNumber.Wrapf("got %d; expected %d", batch.BatchNumber, batchNum)
	}

	err = k.setDataResultTreeEntry(ctx, batchNum, dataEntries)
	if err != nil {
		return 0, err
	}

	for _, valEntry := range valEntries {
		err = k.setValidatorTreeEntry(ctx, batchNum, valEntry)
		if err != nil {
			return 0, err
		}
	}

	_, err = k.incrementCurrentBatchNum(ctx)
	if err != nil {
		return 0, err
	}

	err = k.batches.Set(ctx, batch.BlockHeight, batch)
	if err != nil {
		return 0, err
	}
	return batchNum, nil
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
	return k.GetBatchByBatchNumber(ctx, currentBatchNum-1)
}

// GetLatestSignedBatch returns the latest batch whose signatures have
// been collected.
func (k Keeper) GetLatestSignedBatch(ctx sdk.Context) (types.Batch, error) {
	if ctx.BlockHeight() <= -abci.BlockOffsetCollectPhase {
		return types.Batch{}, types.ErrNoSignedBatch
	}
	currentBatchNum, err := k.currentBatchNumber.Peek(ctx)
	if err != nil {
		return types.Batch{}, err
	}
	if currentBatchNum == collections.DefaultSequenceStart {
		return types.Batch{}, types.ErrBatchingHasNotStarted
	}
	return k.getLatestSignedBatch(ctx, currentBatchNum-1)
}

func (k Keeper) getLatestSignedBatch(ctx sdk.Context, batchNumber uint64) (types.Batch, error) {
	batch, err := k.GetBatchByBatchNumber(ctx, batchNumber)
	if err != nil {
		return types.Batch{}, err
	}
	// If the batch's signatures have not been collected yet, make a
	// recursive call to get the previous batch.
	if batch.BlockHeight > ctx.BlockHeight()+abci.BlockOffsetCollectPhase {
		if batchNumber == collections.DefaultSequenceStart {
			return types.Batch{}, types.ErrNoSignedBatch
		}
		return k.getLatestSignedBatch(ctx, batchNumber-1)
	}
	return batch, nil
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

func (k Keeper) GetValidatorTreeEntry(ctx context.Context, batchNum uint64, valAddr sdk.ValAddress) (types.ValidatorTreeEntry, error) {
	entry, err := k.validatorTreeEntries.Get(ctx, collections.Join(batchNum, valAddr.Bytes()))
	if err != nil {
		return types.ValidatorTreeEntry{}, err
	}
	return entry, nil
}

func (k Keeper) GetValidatorTreeEntries(ctx context.Context, batchNum uint64) ([]types.ValidatorTreeEntry, error) {
	rng := collections.NewPrefixedPairRange[uint64, []byte](batchNum)
	itr, err := k.validatorTreeEntries.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	kvs, err := itr.KeyValues()
	if err != nil {
		return nil, err
	}

	valEntries := make([]types.ValidatorTreeEntry, len(kvs))
	for i, kv := range kvs {
		valEntries[i] = kv.Value
	}
	return valEntries, nil
}

func (k Keeper) GetDataResultTreeEntries(ctx context.Context, batchNum uint64) (types.DataResultTreeEntries, error) {
	dataEntries, err := k.dataResultTreeEntries.Get(ctx, batchNum)
	if err != nil {
		return types.DataResultTreeEntries{}, err
	}
	return dataEntries, nil
}

func (k Keeper) setValidatorTreeEntry(ctx context.Context, batchNum uint64, entry types.ValidatorTreeEntry) error {
	err := k.validatorTreeEntries.Set(ctx, collections.Join(batchNum, entry.ValidatorAddress.Bytes()), entry)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) setDataResultTreeEntry(ctx context.Context, batchNum uint64, dataEntries types.DataResultTreeEntries) error {
	err := k.dataResultTreeEntries.Set(ctx, batchNum, dataEntries)
	if err != nil {
		return err
	}
	return nil
}

// GetBatchData returns various data for a given batch.
func (k Keeper) GetBatchData(ctx context.Context, batchNum uint64) (types.BatchData, error) {
	dataEntries, err := k.GetDataResultTreeEntries(ctx, batchNum)
	if err != nil {
		return types.BatchData{}, err
	}
	valEntries, err := k.GetValidatorTreeEntries(ctx, batchNum)
	if err != nil {
		return types.BatchData{}, err
	}
	sigs, err := k.GetBatchSignatures(ctx, batchNum)
	if err != nil {
		return types.BatchData{}, err
	}
	return types.BatchData{
		BatchNumber:       batchNum,
		DataResultEntries: dataEntries,
		ValidatorEntries:  valEntries,
		BatchSignatures:   sigs,
	}, nil
}

// GetBatchSignatures returns all batch signatures for a given batch.
func (k Keeper) GetBatchSignatures(ctx context.Context, batchNum uint64) ([]types.BatchSignatures, error) {
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

	sigs := make([]types.BatchSignatures, len(kvs))
	for i, kv := range kvs {
		sigs[i] = kv.Value
	}
	return sigs, nil
}

// SetBatchSigSecp256k1 stores a given validator's secp256k1 signature
// for a specified batch.
func (k Keeper) SetBatchSigSecp256k1(ctx context.Context, batchNum uint64, valAddr sdk.ValAddress, signature []byte) error {
	return k.batchSignatures.Set(
		ctx, collections.Join(batchNum, valAddr.Bytes()),
		types.BatchSignatures{
			ValidatorAddress:   valAddr.Bytes(),
			Secp256K1Signature: signature,
		},
	)
}
