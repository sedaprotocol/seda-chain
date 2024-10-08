package keeper

import (
	"context"

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
func (k Keeper) markDataResultAsBatched(ctx context.Context, result types.DataResult) error {
	err := k.dataResults.Remove(ctx, collections.Join(false, result.DrId))
	if err != nil {
		return err
	}
	return k.dataResults.Set(ctx, collections.Join(true, result.DrId), result)
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

// IterateDataResults iterates over all data results under a given
// status (batched or not) and performs a given callback function.
func (k Keeper) IterateDataResults(ctx context.Context, batched bool, cb func(key collections.Pair[bool, string], value types.DataResult) (bool, error)) error {
	rng := collections.NewPrefixedPairRange[bool, string](batched)
	return k.dataResults.Walk(ctx, rng, cb)
}
