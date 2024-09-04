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
	batch, err := q.Keeper.GetBatch(ctx, req.BatchNumber)
	if err != nil {
		return nil, err
	}
	return &types.QueryBatchResponse{
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
