package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) SophonInfo(_ context.Context, _ *types.QuerySophonInfoRequest) (*types.QuerySophonInfoResponse, error) {
	panic("not implemented")
}

func (q Querier) SophonUsers(_ context.Context, _ *types.QuerySophonUsersRequest) (*types.QuerySophonUsersResponse, error) {
	panic("not implemented")
}

func (q Querier) SophonUser(_ context.Context, _ *types.QuerySophonUserRequest) (*types.QuerySophonUserResponse, error) {
	panic("not implemented")
}

func (q Querier) SophonEligibility(_ context.Context, _ *types.QuerySophonEligibilityRequest) (*types.QuerySophonEligibilityResponse, error) {
	panic("not implemented")
}

func (q Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := q.Keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}
