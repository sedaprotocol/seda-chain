package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// SetDataResultForBatching stores a data result so that it is ready
// to be batched.
func (k Keeper) SetDataResultForBatching(ctx context.Context, result types.DataResult) error {
	return k.dataResults.Set(ctx, collections.Join(false, result.DrId), result)
}

// markDataResultAsBatched removes the "unbatched" variant of the given
// data result and stores a "batched" variant.
func (k Keeper) markDataResultAsBatched(ctx context.Context, result types.DataResult, batchNum uint64) error {
	err := k.SetBatchAssignment(ctx, result.DrId, batchNum)
	if err != nil {
		return err
	}
	err = k.dataResults.Remove(ctx, collections.Join(false, result.DrId))
	if err != nil {
		return err
	}
	return k.dataResults.Set(ctx, collections.Join(true, result.DrId), result)
}

// GetDataResult returns a data result given the associated data request's
// ID.
func (k Keeper) GetDataResult(ctx context.Context, dataReqID string) (types.DataResult, error) {
	dataResult, err := k.dataResults.Get(ctx, collections.Join(false, dataReqID))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Look among batched data requests.
			dataResult, err := k.dataResults.Get(ctx, collections.Join(true, dataReqID))
			if err != nil {
				return types.DataResult{}, err
			}
			return dataResult, nil
		}
		return types.DataResult{}, err
	}
	return dataResult, err
}

// GetDataResults returns a list of data results under a given status
// (batched or not).
func (k Keeper) GetDataResults(ctx context.Context, batched bool) ([]types.DataResult, error) {
	var results []types.DataResult
	err := k.IterateDataResults(ctx, batched, func(_ collections.Pair[bool, string], value types.DataResult) (bool, error) {
		results = append(results, value)
		return false, nil
	})
	return results, err
}

// getAllDataResults returns all data results from the store regardless
// of their batched status.
func (k Keeper) getAllDataResults(ctx context.Context) ([]types.DataResult, error) {
	var dataResults []types.DataResult
	unbatched, err := k.GetDataResults(ctx, false)
	if err != nil {
		return nil, err
	}
	dataResults = append(dataResults, unbatched...)
	batched, err := k.GetDataResults(ctx, true)
	if err != nil {
		return nil, err
	}
	dataResults = append(dataResults, batched...)
	return dataResults, nil
}

// IterateDataResults iterates over all data results under a given
// status (batched or not) and performs a given callback function.
func (k Keeper) IterateDataResults(ctx context.Context, batched bool, cb func(key collections.Pair[bool, string], value types.DataResult) (bool, error)) error {
	rng := collections.NewPrefixedPairRange[bool, string](batched)
	return k.dataResults.Walk(ctx, rng, cb)
}

// SetBatchAssignment assigns a given batch number to the given data
// request.
func (k Keeper) SetBatchAssignment(ctx context.Context, dataReqID string, batchNumber uint64) error {
	return k.batchAssignments.Set(ctx, dataReqID, batchNumber)
}

// GetBatchAssignment returns the given data request's assigned batch
// number.
func (k Keeper) GetBatchAssignment(ctx context.Context, dataReqID string) (uint64, error) {
	return k.batchAssignments.Get(ctx, dataReqID)
}

// getAllBatchAssignments retrieves all batch assignments from the store.
func (k Keeper) getAllBatchAssignments(ctx context.Context) ([]types.BatchAssignment, error) {
	var batchAssignments []types.BatchAssignment
	err := k.batchAssignments.Walk(ctx, nil, func(dataReqID string, batchNum uint64) (stop bool, err error) {
		batchAssignments = append(batchAssignments, types.BatchAssignment{
			BatchNumber:   batchNum,
			DataRequestId: dataReqID,
		})
		return false, nil
	})
	return batchAssignments, err
}
