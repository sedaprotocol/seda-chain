package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vrf "github.com/sedaprotocol/vrf-go"

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

func (q Querier) Staker(c context.Context, req *types.QueryStakerRequest) (*types.QueryStakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	staker, err := q.GetStaker(ctx, req.PublicKey)
	if err != nil {
		return nil, err
	}
	return &types.QueryStakerResponse{Staker: staker}, nil
}

func (q Querier) Executors(c context.Context, req *types.QueryExecutorsRequest) (*types.QueryExecutorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	executors, err := q.stakers.GetExecutors(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}
	return &types.QueryExecutorsResponse{Executors: executors}, nil
}

func (q Querier) DataRequestsByStatus(c context.Context, req *types.QueryDataRequestsByStatusRequest) (*types.QueryDataRequestsByStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var lastSeenIndex types.DataRequestIndex
	var err error
	if req.LastSeenIndex != nil {
		lastSeenIndex, err = types.DataRequestIndexFromStrings(req.LastSeenIndex)
		if err != nil {
			return nil, err
		}
	}

	dataRequests, newLastSeenIndex, total, err := q.GetDataRequestsByStatus(ctx, req.Status, req.Limit, lastSeenIndex)
	if err != nil {
		return nil, err
	}

	isPaused, err := q.IsPaused(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataRequestsByStatusResponse{
		DataRequests:  dataRequests,
		IsPaused:      isPaused,
		Total:         total,
		LastSeenIndex: newLastSeenIndex.Strings(),
	}, nil
}

func (q Querier) DataRequestStatuses(c context.Context, req *types.QueryDataRequestStatusesRequest) (*types.QueryDataRequestStatusesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	statuses, err := q.GetDataRequestStatuses(ctx, req.DataRequestIds)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestStatusesResponse{Statuses: statuses}, nil
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

func (q Querier) AccountSeq(c context.Context, req *types.QueryAccountSeqRequest) (*types.QueryAccountSeqResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	staker, err := q.GetStaker(ctx, req.PublicKey)
	if err != nil {
		return nil, err
	}

	return &types.QueryAccountSeqResponse{
		AccountSeq: staker.SequenceNum,
	}, nil
}

func (q Querier) IsStakerExecutor(c context.Context, req *types.QueryIsStakerExecutorRequest) (*types.QueryIsStakerExecutorResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	isExecutor, err := q.Keeper.IsStakerExecutor(ctx, req.PublicKey)
	if err != nil {
		return nil, err
	}
	return &types.QueryIsStakerExecutorResponse{
		IsStakerExecutor: isExecutor,
	}, nil
}

func (q Querier) IsExecutorEligible(c context.Context, req *types.QueryIsExecutorEligibleRequest) (*types.QueryIsExecutorEligibleResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	publicKey, drID, proof, err := req.Parts()
	if err != nil {
		return nil, err
	}

	// Verify the proof.
	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}
	proofBytes, err := hex.DecodeString(proof)
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKeyBytes, proofBytes, req.MsgHash(ctx.ChainID()))
	if err != nil {
		return nil, err
	}

	// Check if the data request exists.
	_, err = q.GetDataRequest(ctx, drID)
	if err != nil {
		return nil, err
	}

	isExecutor, err := q.Keeper.IsStakerExecutor(ctx, publicKey)
	if err != nil {
		return nil, err
	}

	return &types.QueryIsExecutorEligibleResponse{
		IsExecutorEligible: isExecutor,
	}, nil
}
