package keeper

import (
	"context"
	"encoding/hex"
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

func (q Querier) DataRequest(c context.Context, req *types.QueryDataRequestRequest) (*types.QueryDataRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	dataRequest, err := q.GetDataRequest(ctx, req.DrId)
	if err != nil {
		return nil, err
	}
	commits, err := q.GetCommits(ctx, req.DrId)
	if err != nil {
		return nil, err
	}
	reveals, err := q.GetRevealBodies(ctx, req.DrId)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataRequestResponse{
		DataRequest: types.DataRequestResponse{
			DataRequest: dataRequest,
			Commits:     commits,
			Reveals:     reveals,
		},
	}, nil
}

func (q Querier) DataRequestCommitment(c context.Context, req *types.QueryDataRequestCommitmentRequest) (*types.QueryDataRequestCommitmentResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	commitmentBytes, err := q.GetCommit(ctx, req.DrId, req.PublicKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, types.ErrNotCommitted
		}
		return nil, err
	}
	return &types.QueryDataRequestCommitmentResponse{Commitment: hex.EncodeToString(commitmentBytes)}, nil
}

func (q Querier) DataRequestCommitments(c context.Context, req *types.QueryDataRequestCommitmentsRequest) (*types.QueryDataRequestCommitmentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	commits, err := q.GetCommitsHexEncoded(ctx, req.DrId)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestCommitmentsResponse{Commitments: commits}, nil
}

func (q Querier) DataRequestReveal(c context.Context, req *types.QueryDataRequestRevealRequest) (*types.QueryDataRequestRevealResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	revealBody, err := q.GetRevealBody(ctx, req.DrId, req.PublicKey)
	if err != nil {
		return nil, err
	}

	return &types.QueryDataRequestRevealResponse{Reveal: &revealBody}, nil
}

func (q Querier) DataRequestReveals(c context.Context, req *types.QueryDataRequestRevealsRequest) (*types.QueryDataRequestRevealsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	reveals, err := q.GetRevealBodies(ctx, req.DrId)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestRevealsResponse{Reveals: reveals}, nil
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

	ids, newLastSeenIndex, total, err := q.dataRequests.GetDataRequestIDsByStatusPaginated(ctx, req.Status, req.Limit, lastSeenIndex)
	if err != nil {
		return nil, err
	}

	dataRequests := make([]types.DataRequestResponse, 0, len(ids))
	for _, id := range ids {
		dataRequest, err := q.GetDataRequest(ctx, id)
		if err != nil {
			return nil, err
		}
		commits, err := q.GetCommits(ctx, id)
		if err != nil {
			return nil, err
		}
		reveals, err := q.GetRevealBodies(ctx, id)
		if err != nil {
			return nil, err
		}

		dataRequests = append(dataRequests, types.DataRequestResponse{
			DataRequest: dataRequest,
			Commits:     commits,
			Reveals:     reveals,
		})
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
		// An account that does not exist has a sequence number of 0.
		if errors.Is(err, collections.ErrNotFound) {
			return &types.QueryAccountSeqResponse{
				AccountSeq: 0,
			}, nil
		}
		return nil, err
	}

	return &types.QueryAccountSeqResponse{
		AccountSeq: staker.SequenceNum,
	}, nil
}

func (q Querier) StakerAndSeq(c context.Context, req *types.QueryStakerAndSeqRequest) (*types.QueryStakerAndSeqResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	staker, err := q.GetStaker(ctx, req.PublicKey)
	if err != nil {
		// An account that does not exist has a sequence number of 0.
		if errors.Is(err, collections.ErrNotFound) {
			return &types.QueryStakerAndSeqResponse{}, nil
		}
		return nil, err
	}
	return &types.QueryStakerAndSeqResponse{Staker: staker, SequenceNum: staker.SequenceNum}, nil
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

func (q Querier) ExecutorEligibility(c context.Context, req *types.QueryExecutorEligibilityRequest) (*types.QueryExecutorEligibilityResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	status, errMsg, err := q.GetExecutorEligibility(ctx, req)
	if err != nil {
		return nil, err
	}
	return &types.QueryExecutorEligibilityResponse{
		//nolint:gosec // G115: Block height should never be negative.
		BlockHeight:  uint64(ctx.BlockHeight()),
		Status:       status,
		ErrorMessage: errMsg,
	}, nil
}

func (q Querier) LegacyExecutorEligibility(c context.Context, req *types.QueryLegacyExecutorEligibilityRequest) (*types.QueryLegacyExecutorEligibilityResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	status, errMsg, err := q.GetExecutorEligibility(ctx, req)
	if err != nil {
		return nil, err
	}
	return &types.QueryLegacyExecutorEligibilityResponse{
		//nolint:gosec // G115: Block height should never be negative.
		BlockHeight:  uint64(ctx.BlockHeight()),
		Status:       status,
		ErrorMessage: errMsg,
	}, nil
}
