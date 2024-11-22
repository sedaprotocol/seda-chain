package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sedaprotocol/seda-chain/app/abci"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerierImpl(keeper Keeper) types.QueryServer {
	return &Querier{
		keeper,
	}
}

func (q Querier) Batch(c context.Context, req *types.QueryBatchRequest) (*types.QueryBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var batch types.Batch
	var err error
	if req.BatchNumber == 0 {
		batch, err = q.Keeper.GetBatchForHeight(ctx, ctx.BlockHeight()+abci.BlockOffsetCollectPhase)
	} else {
		batch, err = q.Keeper.GetBatchByBatchNumber(ctx, req.BatchNumber)
	}
	if err != nil {
		return nil, err
	}

	return &types.QueryBatchResponse{
		Batch: batch,
	}, nil
}

func (q Querier) BatchForHeight(c context.Context, req *types.QueryBatchForHeightRequest) (*types.QueryBatchForHeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	batch, err := q.Keeper.GetBatchForHeight(ctx, req.BlockHeight)
	if err != nil {
		return nil, err
	}
	return &types.QueryBatchForHeightResponse{
		Batch: batch,
	}, nil
}

func (q Querier) Batches(c context.Context, req *types.QueryBatchesRequest) (*types.QueryBatchesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	batches, pageRes, err := query.CollectionFilteredPaginate(
		ctx, q.batches, req.Pagination,
		func(_ int64, value types.Batch) (bool, error) {
			if value.BlockHeight > ctx.BlockHeight()+abci.BlockOffsetCollectPhase {
				return false, nil
			}
			return true, nil
		},
		func(_ int64, value types.Batch) (types.Batch, error) {
			return value, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryBatchesResponse{
		Batches:    batches,
		Pagination: pageRes,
	}, nil
}

func (q Querier) TreeEntries(c context.Context, req *types.QueryTreeEntriesRequest) (*types.QueryTreeEntriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	entries, err := q.Keeper.GetTreeEntriesForBatch(ctx, req.BatchNumber)
	if err != nil {
		return nil, err
	}
	return &types.QueryTreeEntriesResponse{
		Entries: entries,
	}, nil
}

func (q Querier) DataResult(c context.Context, req *types.QueryDataResultRequest) (*types.QueryDataResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	dataResult, err := q.Keeper.GetDataResult(ctx, req.DataRequestId)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataResultResponse{
		DataResult: dataResult,
	}, nil
}

func (q Querier) BatchAssignment(c context.Context, req *types.QueryBatchAssignmentRequest) (*types.QueryBatchAssignmentResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	batchNum, err := q.Keeper.GetBatchAssignment(ctx, req.DataRequestId)
	if err != nil {
		return nil, err
	}
	return &types.QueryBatchAssignmentResponse{
		BatchNumber: batchNum,
	}, nil
}
