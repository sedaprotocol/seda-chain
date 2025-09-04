package keeper

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"

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

func (q Querier) SophonEligibility(c context.Context, req *types.QuerySophonEligibilityRequest) (*types.QuerySophonEligibilityResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	payload, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid base64 in payload: %s", req.Payload)
	}

	// The format is "{blockNumber}:{userId}:{signature_hex_string}"
	parts := strings.Split(string(payload), ":")
	if len(parts) != 3 {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid number of parts: %s", string(payload))
	}

	blockHeight, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid block height: %s", parts[0])
	}

	userID := parts[1]

	signature := parts[2]
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in signature: %s", signature)
	}

	// The signed hash is "blocknumber_be_uint64, keccak256(userId_utf8_bytes), chainId_utf8_bytes"
	var payloadBytes []byte
	payloadBytes = binary.BigEndian.AppendUint64(payloadBytes, blockHeight)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(userID))
	userIDHash := hasher.Sum(nil)
	payloadBytes = append(payloadBytes, userIDHash...)

	payloadBytes = append(payloadBytes, []byte(ctx.ChainID())...)

	// Make the hash that should have been signed
	hasher.Reset()
	hasher.Write(payloadBytes)
	payloadHash := hasher.Sum(nil)

	sigPubKey, err := crypto.Ecrecover(payloadHash, signatureBytes)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid signature: %s", signature)
	}

	sophonInfo, err := q.GetSophonInfo(ctx, sigPubKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no sophon registered for %s", hex.EncodeToString(sigPubKey))
		}

		return nil, err
	}

	sophonUser, err := q.GetSophonUser(ctx, sophonInfo.Id, userID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no user registered for %s", userID)
		}

		return nil, err
	}

	// The querying service is responsible for determining whether they want to handle the request or not,
	// this query simply returns whether the signature is valid, the credits of the user, and the current block height.
	return &types.QuerySophonEligibilityResponse{
		Eligible:    true,
		UserCredits: sophonUser.Credits,
		//nolint:gosec // G115: We shouldn't get negative block heights
		BlockHeight: uint64(ctx.BlockHeight()),
	}, nil
}

func (q Querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params, err := q.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}
