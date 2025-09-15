package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) Allowlist(c context.Context, _ *types.QueryAllowlistRequest) (*types.QueryAllowlistResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	iter, _ := q.allowlist.Iterate(ctx, nil)
	defer iter.Close()
	keys, _ := iter.Keys()

	return &types.QueryAllowlistResponse{
		PublicKeys: keys,
	}, nil
}

func (q Querier) StakerAndSeq(c context.Context, req *types.QueryStakerAndSeqRequest) (*types.QueryStakerAndSeqResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	staker, err := q.GetStaker(ctx, req.PublicKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &types.QueryStakerAndSeqResponse{}, nil
		}
		return nil, err
	}
	return &types.QueryStakerAndSeqResponse{Staker: staker, SequenceNum: staker.SequenceNum}, nil
}

func (q Querier) Paused(c context.Context, _ *types.QueryPausedRequest) (*types.QueryPausedResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	paused, err := q.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryPausedResponse{Paused: paused}, nil
}

func (q Querier) PendingOwner(c context.Context, _ *types.QueryPendingOwnerRequest) (*types.QueryPendingOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pendingOwner, err := q.GetPendingOwner(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryPendingOwnerResponse{PendingOwner: pendingOwner}, nil
}

func (q Querier) Owner(c context.Context, _ *types.QueryOwnerRequest) (*types.QueryOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	owner, err := q.GetOwner(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryOwnerResponse{Owner: owner}, nil
}

func (q Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := q.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) StakingConfig(c context.Context, _ *types.QueryStakingConfigRequest) (*types.QueryStakingConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	stakingConfig, err := q.GetStakingConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryStakingConfigResponse{StakingConfig: stakingConfig}, nil
}

func (q Querier) DataRequestConfig(c context.Context, _ *types.QueryDataRequestConfigRequest) (*types.QueryDataRequestConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	dataRequestConfig, err := q.GetDataRequestConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestConfigResponse{DataRequestConfig: dataRequestConfig}, nil
}
