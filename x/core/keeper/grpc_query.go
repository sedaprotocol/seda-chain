package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/sha3"

	"github.com/sedaprotocol/seda-chain/x/core/types"
	vrf "github.com/sedaprotocol/vrf-go"
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

func isExecutorEligibleMsgHash(publicKeyBytes, drIdBytes []byte) []byte {
	allBytes := append([]byte{}, []byte("is_executor_eligible")...)
	allBytes = append(allBytes, drIdBytes...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func isExecutorEligibleLegacyMsgHash(contractAddr string, publicKeyBytes, drIdBytes []byte) []byte {
	allBytes := append([]byte{}, []byte("is_executor_eligible")...)
	allBytes = append(allBytes, drIdBytes...)
	allBytes = append(allBytes, []byte(contractAddr)...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (q Querier) IsExecutorEligible(c context.Context, req *types.QueryIsExecutorEligibleRequest) (*types.QueryIsExecutorEligibleResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// verify the proof
	hash := req.MsgHash()
	publicKeyBytes, drIDBytes, proof, err := req.Parts()
	if err != nil {
		return nil, err
	}
	vrf.NewK256VRF().Verify(publicKeyBytes, proof, hash)

	// TODO: we should store the drID as bytes to avoid this
	hexDrID := hex.EncodeToString(drIDBytes)
	// check if dr is in the data request pool
	_, err = q.GetDataRequest(ctx, hexDrID)
	if err != nil {
		return nil, err
	}

	// TODO: move IsStakerExecutor logic to a separate function and call it here
	isExecutor, err := q.Keeper.IsStakerExecutor(ctx, hex.EncodeToString(publicKeyBytes))
	if err != nil {
		return nil, err
	}
	if !isExecutor {
		return &types.QueryIsExecutorEligibleResponse{
			IsExecutorEligible: false,
		}, nil
	}

	return &types.QueryIsExecutorEligibleResponse{
		IsExecutorEligible: false,
	}, nil
}
