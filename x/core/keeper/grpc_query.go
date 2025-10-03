package keeper

import (
	"context"
	"encoding/hex"
	"errors"

	"cosmossdk.io/collections"

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

func (q Querier) DataRequest(c context.Context, req *types.QueryDataRequestRequest) (*types.QueryDataRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	dataRequest, err := q.GetDataRequest(ctx, req.DrId)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestResponse{DataRequest: dataRequest}, nil
}

func (q Querier) DataRequestCommitment(c context.Context, req *types.QueryDataRequestCommitmentRequest) (*types.QueryDataRequestCommitmentResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	dataRequest, err := q.GetDataRequest(ctx, req.DrId)
	if err != nil {
		return nil, err
	}

	commitmentBytes, exists := dataRequest.GetCommit(req.PublicKey)
	if !exists {
		return nil, types.ErrNotCommitted
	}
	commitment := hex.EncodeToString(commitmentBytes)

	return &types.QueryDataRequestCommitmentResponse{Commitment: commitment}, nil
}

func (q Querier) DataRequestCommitments(c context.Context, req *types.QueryDataRequestCommitmentsRequest) (*types.QueryDataRequestCommitmentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	dataRequest, err := q.GetDataRequest(ctx, req.DrId)
	if err != nil {
		return nil, err
	}

	commitments := make(map[string]string, len(dataRequest.Commits))
	for pubKey, commitBytes := range dataRequest.Commits {
		commitments[pubKey] = hex.EncodeToString(commitBytes)
	}

	return &types.QueryDataRequestCommitmentsResponse{Commitments: commitments}, nil
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
	dataRequest, err := q.GetDataRequest(ctx, req.DrId)
	if err != nil {
		return nil, err
	}

	reveals := make(map[string]*types.RevealBody, len(dataRequest.Reveals))
	for pubKey := range dataRequest.Reveals {
		revealBody, err := q.GetRevealBody(ctx, req.DrId, pubKey)
		if err != nil {
			return nil, err
		}
		reveals[pubKey] = &revealBody
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

func (q Querier) GetExecutorEligibility(c context.Context, req *types.QueryGetExecutorEligibilityRequest) (*types.QueryGetExecutorEligibilityResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	result := &types.QueryGetExecutorEligibilityResponse{
		//nolint:gosec // G115: Block height should never be negative.
		BlockHeight: uint64(ctx.BlockHeight()),
	}

	publicKey, drID, proof, err := req.Parts()
	if err != nil {
		return nil, err
	}

	// Check the executor's status as a staker.
	staker, err := q.GetStaker(ctx, publicKey)
	if err != nil {
		result.Status = types.ELIGIBILITY_STATUS_NOT_STAKER
		result.ErrorMessage = err.Error()
		return result, nil
	}

	stakingConfig, err := q.GetStakingConfig(ctx)
	if err != nil {
		return nil, err
	}
	if stakingConfig.AllowlistEnabled {
		isAllowlisted, err := q.IsAllowlisted(ctx, publicKey)
		if err != nil {
			return nil, err
		}
		if !isAllowlisted {
			result.Status = types.ELIGIBILITY_STATUS_NOT_ALLOWLISTED
			return result, nil
		}
	}

	if staker.Staked.LT(stakingConfig.MinimumStake) {
		result.Status = types.ELIGIBILITY_STATUS_INSUFFICIENT_STAKE
		return result, nil
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
		result.Status = types.ELIGIBILITY_STATUS_INVALID_SIGNATURE
		result.ErrorMessage = err.Error()
		return result, nil
	}

	// Verify eligibility with respect to the data request.
	drIDBytes, err := hex.DecodeString(drID)
	if err != nil {
		return nil, err
	}

	isEligible, err := q.IsEligibleForDataRequest(ctx, publicKeyBytes, drIDBytes, stakingConfig.MinimumStake)
	if err != nil {
		result.Status = types.ELIGIBILITY_STATUS_NOT_ELIGIBLE
		result.ErrorMessage = err.Error()
		return result, nil
	}

	if isEligible {
		result.Status = types.ELIGIBILITY_STATUS_ELLIGIBLE
	} else {
		result.Status = types.ELIGIBILITY_STATUS_NOT_ELIGIBLE
	}
	return result, nil
}
