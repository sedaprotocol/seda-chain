package keeper

import (
	"context"
	"encoding/hex"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

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

	return &types.QuerySophonInfoResponse{Info: result}, nil
}

func (q Querier) SophonTransfer(c context.Context, req *types.QuerySophonTransferRequest) (*types.QuerySophonTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pubKeyBytes, err := hex.DecodeString(req.SophonPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.SophonPubKey)
	}

	sophonInfo, err := q.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no sophon registered for %s", req.SophonPubKey)
		}

		return nil, err
	}

	transferAddress, err := q.GetSophonTransfer(ctx, sophonInfo.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no sophon transfer pending for %s", req.SophonPubKey)
		}

		return nil, err
	}

	newOwnerAddress := sdk.AccAddress(transferAddress).String()

	return &types.QuerySophonTransferResponse{NewOwnerAddress: newOwnerAddress}, nil
}

func (q Querier) SophonUsers(c context.Context, req *types.QuerySophonUsersRequest) (*types.QuerySophonUsersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pubKeyBytes, err := hex.DecodeString(req.SophonPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.SophonPubKey)
	}

	sophonInfo, err := q.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no sophon registered for %s", req.SophonPubKey)
		}

		return nil, err
	}

	users, pageRes, err := query.CollectionPaginate(
		ctx, q.sophonUser, req.Pagination,
		func(_ collections.Pair[uint64, string], value types.SophonUser) (types.SophonUser, error) {
			return value, nil
		},
		func(opts *query.CollectionsPaginateOptions[collections.Pair[uint64, string]]) {
			prefix := collections.PairPrefix[uint64, string](sophonInfo.Id)
			opts.Prefix = &prefix
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.QuerySophonUsersResponse{
		Users:      users,
		Pagination: pageRes,
	}, nil
}

func (q Querier) SophonUser(c context.Context, req *types.QuerySophonUserRequest) (*types.QuerySophonUserResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pubKeyBytes, err := hex.DecodeString(req.SophonPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.SophonPubKey)
	}

	sophonInfo, err := q.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no sophon registered for %s", req.SophonPubKey)
		}

		return nil, err
	}

	user, err := q.GetSophonUser(ctx, sophonInfo.Id, req.UserId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no user registered for %s", req.UserId)
		}

		return nil, err
	}

	return &types.QuerySophonUserResponse{User: user}, nil
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
