package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// SetBatchAssignment stores mapping between data request ID - posted height pair
// and assigned batch number in both directions.
func (k Keeper) SetBatchAssignment(ctx context.Context, dataReqID string, dataReqHeight, batchNumber uint64) error {
	items, err := k.batchDataResults.Get(ctx, batchNumber)
	if err != nil {
		if !errors.Is(err, collections.ErrNotFound) {
			return err
		}
		items.DataRequestIdHeights = make([]types.DataRequestIDHeight, 0)
	}
	items.DataRequestIdHeights = append(items.DataRequestIdHeights, types.DataRequestIDHeight{
		DataRequestId:     dataReqID,
		DataRequestHeight: dataReqHeight,
	})
	err = k.batchDataResults.Set(ctx, batchNumber, items)
	if err != nil {
		return err
	}

	return k.batchAssignments.Set(ctx, collections.Join(dataReqID, dataReqHeight), batchNumber)
}

func (k Keeper) GetBatchAssignment(ctx context.Context, dataReqID string, dataReqHeight uint64) (uint64, error) {
	return k.batchAssignments.Get(ctx, collections.Join(dataReqID, dataReqHeight))
}

func (k Keeper) RemoveBatchAssignment(ctx context.Context, dataReqID string, dataReqHeight uint64) error {
	return k.batchAssignments.Remove(ctx, collections.Join(dataReqID, dataReqHeight))
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

func (k Keeper) SetBatchDataResults(ctx context.Context, batchNumber uint64, dataRequestIDHeights types.DataRequestIDHeights) error {
	return k.batchDataResults.Set(ctx, batchNumber, dataRequestIDHeights)
}

func (k Keeper) GetBatchDataResults(ctx context.Context, batchNumber uint64) (types.DataRequestIDHeights, error) {
	return k.batchDataResults.Get(ctx, batchNumber)
}

func (k Keeper) RemoveBatchDataResults(ctx context.Context, batchNumber uint64) error {
	return k.batchDataResults.Remove(ctx, batchNumber)
}
