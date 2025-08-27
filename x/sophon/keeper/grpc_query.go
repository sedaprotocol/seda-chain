package keeper

import (
	"context"
	"encoding/hex"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) SophonInfo(c context.Context, req *types.QuerySophonInfoRequest) (*types.QuerySophonInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pubKeyBytes, err := hex.DecodeString(req.SophonPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.SophonPubKey)
	}

	result, err := q.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no sophon registered for %s", req.SophonPubKey)
		}

		return nil, err
	}

	return &types.QuerySophonInfoResponse{Info: &result}, nil
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
