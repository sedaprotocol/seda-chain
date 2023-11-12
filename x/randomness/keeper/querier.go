package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
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

func (q Querier) Seed(c context.Context, req *types.QuerySeedRequest) (*types.QuerySeedResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QuerySeedResponse{
		Seed:        q.GetSeed(ctx),
		BlockHeight: ctx.BlockHeight(),
	}, nil
}
