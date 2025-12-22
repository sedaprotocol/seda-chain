package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// SetDataResultForBatching stores a data result so that it is ready to be batched.
func (k Keeper) SetDataResultForBatching(ctx context.Context, result types.DataResult) error {
	return k.dataResults.Set(ctx, collections.Join3(false, result.DrId, result.DrBlockHeight), result)
}

// SetDataResultAsBatched stores a data result under "batched" status.
func (k Keeper) SetDataResultAsBatched(ctx context.Context, result types.DataResult) error {
	return k.dataResults.Set(ctx, collections.Join3(true, result.DrId, result.DrBlockHeight), result)
}

// RemoveDataResult removes a data result from the store.
func (k Keeper) RemoveDataResult(ctx context.Context, batched bool, dataReqID string, dataReqHeight uint64) error {
	return k.dataResults.Remove(ctx, collections.Join3(batched, dataReqID, dataReqHeight))
}

// RemoveLegacyDataResult removes a data result from the legacy store.
func (k Keeper) RemoveLegacyDataResult(ctx context.Context, batched bool, dataReqID string, dataReqHeight uint64) error {
	return k.legacyDataResults.Remove(ctx, collections.Join3(batched, dataReqID, dataReqHeight))
}

// MarkDataResultAsBatched updates the data result status to "batched" and stores
// the batch number assignment.
func (k Keeper) MarkDataResultAsBatched(ctx context.Context, result types.DataResult, batchNum uint64) error {
	err := k.RemoveDataResult(ctx, false, result.DrId, result.DrBlockHeight)
	if err != nil {
		return err
	}
	err = k.SetDataResultAsBatched(ctx, result)
	if err != nil {
		return err
	}
	return k.SetBatchAssignment(ctx, result.DrId, result.DrBlockHeight, batchNum)
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

// IterateDataResults iterates over all data results under a given
// status (batched or not) and performs a given callback function.
func (k Keeper) IterateDataResults(ctx context.Context, batched bool, cb func(key collections.Triple[bool, string, uint64], value types.DataResult) (bool, error)) error {
	rng := collections.NewPrefixedTripleRange[bool, string, uint64](batched)
	return k.dataResults.Walk(ctx, rng, cb)
}

// getAllGenesisDataResults returns all data results from the store regardless
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

// GetLegacyDataResults returns a list of legacy data results under a given status
// (batched or not).
func (k Keeper) GetLegacyDataResults(ctx context.Context, batched bool) ([]types.DataResult, error) {
	var results []types.DataResult
	err := k.IterateLegacyDataResults(ctx, batched, func(_ collections.Triple[bool, string, uint64], value types.DataResult) (bool, error) {
		results = append(results, value)
		return false, nil
	})
	return results, err
}

// IterateLegacyDataResults iterates over all legacy data results under a given
// status (batched or not) and performs a given callback function.
func (k Keeper) IterateLegacyDataResults(ctx context.Context, batched bool, cb func(key collections.Triple[bool, string, uint64], value types.DataResult) (bool, error)) error {
	rng := collections.NewPrefixedTripleRange[bool, string, uint64](batched)
	return k.legacyDataResults.Walk(ctx, rng, cb)
}

// getAllGenesisLegacyDataResults returns all data results from the legacy store regardless
// of their batched status. Used for genesis export.
func (k Keeper) getAllGenesisLegacyDataResults(ctx context.Context) ([]types.GenesisDataResult, error) {
	dataResults := make([]types.GenesisDataResult, 0)
	unbatched, err := k.GetLegacyDataResults(ctx, false)
	if err != nil {
		return nil, err
	}
	for _, result := range unbatched {
		dataResults = append(dataResults, types.GenesisDataResult{
			Batched:    false,
			DataResult: result,
		})
	}
	batched, err := k.GetLegacyDataResults(ctx, true)
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

// GetDataResult returns a data result given the associated data request's
// ID and height.
// NOTE: Checks both legacy and new collections.
func (k Keeper) GetDataResult(ctx context.Context, dataReqID string, dataReqHeight uint64) (types.DataResult, error) {
	var dataResult types.DataResult
	var err error

	// Check in the following order:
	// Legacy unbatched -> Legacy batched -> Unbatched -> Batched
	dataResult, err = k.legacyDataResults.Get(ctx, collections.Join3(false, dataReqID, dataReqHeight))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return dataResult, err
	} else if err == nil {
		return dataResult, nil
	}
	dataResult, err = k.legacyDataResults.Get(ctx, collections.Join3(true, dataReqID, dataReqHeight))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return dataResult, err
	} else if err == nil {
		return dataResult, nil
	}

	dataResult, err = k.dataResults.Get(ctx, collections.Join3(false, dataReqID, dataReqHeight))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return dataResult, err
	} else if err == nil {
		return dataResult, nil
	}
	return k.dataResults.Get(ctx, collections.Join3(true, dataReqID, dataReqHeight))
}

// GetLatestDataResult returns the latest data result given the associated
// data request's ID.
// NOTE: Checks both legacy and new collections.
func (k Keeper) GetLatestDataResult(ctx context.Context, dataReqID string) (types.DataResult, error) {
	dataResult, err := k.getLatestDataResult(ctx, false, dataReqID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			// Look among batched data requests.
			dataResult, err := k.getLatestDataResult(ctx, true, dataReqID)
			if err != nil {
				return types.DataResult{}, err
			}
			return dataResult, nil
		}
		return types.DataResult{}, err
	}
	return dataResult, nil
}

// getLatestDataResult returns the latest data result given the associated
// data request's ID and its batched status.
// NOTE: Checks both legacy and new collections.
func (k Keeper) getLatestDataResult(ctx context.Context, batched bool, dataReqID string) (types.DataResult, error) {
	// The triple pair ranger does not expose the Descending() method,
	// so we manually create the range using the same prefix that the
	// collections.NewSuperPrefixedTripleRange uses internally.
	drRange := &collections.Range[collections.Triple[bool, string, uint64]]{}
	drRange.Prefix(collections.TripleSuperPrefix[bool, string, uint64](batched, dataReqID)).Descending()

	// Check the current store first.
	itr, err := k.dataResults.Iterate(ctx, drRange)
	if err != nil {
		return types.DataResult{}, err
	}
	defer itr.Close()

	if itr.Valid() {
		kv, err := itr.KeyValue()
		if err != nil {
			return types.DataResult{}, err
		}
		return kv.Value, nil
	}

	// Check the legacy store.
	itr, err = k.legacyDataResults.Iterate(ctx, drRange)
	if err != nil {
		return types.DataResult{}, err
	}
	defer itr.Close()

	if itr.Valid() {
		kv, err := itr.KeyValue()
		if err != nil {
			return types.DataResult{}, err
		}
		return kv.Value, nil
	}

	return types.DataResult{}, collections.ErrNotFound
}
