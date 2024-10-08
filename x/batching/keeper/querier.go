package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	batch, err := q.Keeper.GetBatchByBatchNumber(ctx, req.BatchNumber)
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

func (q Querier) Batches(c context.Context, _ *types.QueryBatchesRequest) (*types.QueryBatchesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	batches, err := q.GetAllBatches(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryBatchesResponse{
		Batches: batches,
	}, nil
}

func (q Querier) TreeEntries(c context.Context, req *types.QueryTreeEntriesRequest) (*types.QueryTreeEntriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	entries, err := q.Keeper.GetTreeEntries(ctx, req.BatchNumber)
	if err != nil {
		return nil, err
	}
	return &types.QueryTreeEntriesResponse{
		Entries: entries,
	}, nil
}
