package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := q.Keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) ValidatorKeys(ctx context.Context, req *types.QueryValidatorKeysRequest) (*types.QueryValidatorKeysResponse, error) {
	result, err := q.GetValidatorKeys(ctx, req.ValidatorAddr)
	if err != nil {
		return nil, err
	}
	return &types.QueryValidatorKeysResponse{ValidatorPubKeys: result}, nil
}

func (q Querier) ProvingSchemes(ctx context.Context, _ *types.QueryProvingSchemesRequest) (*types.QueryProvingSchemesResponse, error) {
	schemes, err := q.GetAllProvingSchemes(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return nil, err
	}
	return &types.QueryProvingSchemesResponse{
		ProvingSchemes: schemes,
	}, nil
}
