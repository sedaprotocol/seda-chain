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
	return k.dataResults.Set(ctx, collections.Join3(false, result.DrId, result.DrBlockHeight), result)
}

// MarkDataResultAsBatched removes the "unbatched" variant of the given
// data result and stores a "batched" variant.
func (k Keeper) MarkDataResultAsBatched(ctx context.Context, result types.DataResult, batchNum uint64) error {
	err := k.SetBatchAssignment(ctx, result.DrId, result.DrBlockHeight, batchNum)
	if err != nil {
		return err
	}
	err = k.dataResults.Remove(ctx, collections.Join3(false, result.DrId, result.DrBlockHeight))
	if err != nil {
		return err
	}
	return k.dataResults.Set(ctx, collections.Join3(true, result.DrId, result.DrBlockHeight), result)
}

// GetDataResult returns a data result given the associated data request's
// ID and height.
func (k Keeper) GetDataResult(ctx context.Context, dataReqID string, dataReqHeight uint64) (*types.DataResult, error) {
	dataResult, err := k.dataResults.Get(ctx, collections.Join3(false, dataReqID, dataReqHeight))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Look among batched data requests.
			dataResult, err := k.dataResults.Get(ctx, collections.Join3(true, dataReqID, dataReqHeight))
			if err != nil {
				return nil, err
			}
			return &dataResult, nil
		}
		return nil, err
	}
	return &dataResult, err
}

// GetLatestDataResult returns the latest data result given the associated
// data request's ID.
func (k Keeper) GetLatestDataResult(ctx context.Context, dataReqID string) (*types.DataResult, error) {
	dataResult, err := k.getLatestDataResult(ctx, false, dataReqID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Look among batched data requests.
			dataResult, err := k.getLatestDataResult(ctx, true, dataReqID)
			if err != nil {
				return nil, err
			}
			return dataResult, nil
		}
		return nil, err
	}

	return dataResult, nil
}

func (k Keeper) getLatestDataResult(ctx context.Context, batched bool, dataReqID string) (*types.DataResult, error) {
	// The triple pair ranger does not expose the Descending() method,
	// so we manually create the range using the same prefix that the
	// collections.NewSuperPrefixedTripleRange uses internally.
	drRange := &collections.Range[collections.Triple[bool, string, uint64]]{}
	drRange.Prefix(collections.TripleSuperPrefix[bool, string, uint64](batched, dataReqID)).Descending()

	itr, err := k.dataResults.Iterate(ctx, drRange)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	if itr.Valid() {
		kv, err := itr.KeyValue()
		if err != nil {
			return nil, err
		}
		return &kv.Value, nil
	}

	return nil, collections.ErrNotFound
}

// GetDataResults returns a list of data results under a given status
// (batched or not).
func (k Keeper) GetDataResults(ctx context.Context, batched bool) ([]types.DataResult, error) {
	var results []types.DataResult
	err := k.IterateDataResults(ctx, batched, func(_ collections.Triple[bool, string, uint64], value types.DataResult) (bool, error) {
		results = append(results, value)
		return false, nil
	})
	return results, err
}

// getAllDataResults returns all data results from the store regardless
// of their batched status. Used for genesis export.
func (k Keeper) getAllGenesisDataResults(ctx context.Context) ([]types.GenesisDataResult, error) {
	dataResults := make([]types.GenesisDataResult, 0)
	unbatched, err := k.GetDataResults(ctx, false)
	if err != nil {
		return nil, err
	}
	for _, result := range unbatched {
		dataResults = append(dataResults, types.GenesisDataResult{
			Batched:    false,
			DataResult: result,
		})
	}
	batched, err := k.GetDataResults(ctx, true)
	if err != nil {
		return nil, err
	}
	for _, result := range batched {
		dataResults = append(dataResults, types.GenesisDataResult{
			Batched:    true,
			DataResult: result,
		})
	}
	return dataResults, nil
}

// IterateDataResults iterates over all data results under a given
// status (batched or not) and performs a given callback function.
func (k Keeper) IterateDataResults(ctx context.Context, batched bool, cb func(key collections.Triple[bool, string, uint64], value types.DataResult) (bool, error)) error {
	rng := collections.NewPrefixedTripleRange[bool, string, uint64](batched)
	return k.dataResults.Walk(ctx, rng, cb)
}

// SetBatchAssignment assigns a given batch number to the given data
// request ID and data request height.
func (k Keeper) SetBatchAssignment(ctx context.Context, dataReqID string, dataReqHeight uint64, batchNumber uint64) error {
	return k.batchAssignments.Set(ctx, collections.Join(dataReqID, dataReqHeight), batchNumber)
}

// GetBatchAssignment returns the given data request's assigned batch
// number for a given height.
func (k Keeper) GetBatchAssignment(ctx context.Context, dataReqID string, dataReqHeight uint64) (uint64, error) {
	return k.batchAssignments.Get(ctx, collections.Join(dataReqID, dataReqHeight))
}

// getAllBatchAssignments retrieves all batch assignments from the store.
// Used for genesis export.
func (k Keeper) getAllBatchAssignments(ctx context.Context) ([]types.BatchAssignment, error) {
	var batchAssignments []types.BatchAssignment
	err := k.batchAssignments.Walk(ctx, nil, func(key collections.Pair[string, uint64], value uint64) (stop bool, err error) {
		batchAssignments = append(batchAssignments, types.BatchAssignment{
			BatchNumber:       value,
			DataRequestId:     key.K1(),
			DataRequestHeight: key.K2(),
		})
		return false, nil
	})
	return batchAssignments, err
}
