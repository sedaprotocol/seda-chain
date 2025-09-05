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

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func (q Querier) FastClient(c context.Context, req *types.QueryFastClientRequest) (*types.QueryFastClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pubKeyBytes, err := hex.DecodeString(req.FastClientPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.FastClientPubKey)
	}

	result, err := q.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no fast client registered for %s", req.FastClientPubKey)
		}

		return nil, err
	}

	return &types.QueryFastClientResponse{Client: result}, nil
}

func (q Querier) FastClientTransfer(c context.Context, req *types.QueryFastClientTransferRequest) (*types.QueryFastClientTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	pubKeyBytes, err := hex.DecodeString(req.FastClientPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.FastClientPubKey)
	}

	fastClient, err := q.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no fast client registered for %s", req.FastClientPubKey)
		}

		return nil, err
	}

	transferAddress, err := q.GetFastTransfer(ctx, fastClient.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no fast client transfer pending for %s", req.FastClientPubKey)
		}

		return nil, err
	}

	newOwnerAddress := sdk.AccAddress(transferAddress).String()

	return &types.QueryFastClientTransferResponse{NewOwnerAddress: newOwnerAddress}, nil
}

func (q Querier) FastClientUsers(c context.Context, req *types.QueryFastClientUsersRequest) (*types.QueryFastClientUsersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pubKeyBytes, err := hex.DecodeString(req.FastClientPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.FastClientPubKey)
	}

	fastClient, err := q.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no fast client registered for %s", req.FastClientPubKey)
		}

		return nil, err
	}

	users, pageRes, err := query.CollectionPaginate(
		ctx, q.fastUser, req.Pagination,
		func(_ collections.Pair[uint64, string], value types.FastUser) (types.FastUser, error) {
			return value, nil
		},
		func(opts *query.CollectionsPaginateOptions[collections.Pair[uint64, string]]) {
			prefix := collections.PairPrefix[uint64, string](fastClient.Id)
			opts.Prefix = &prefix
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryFastClientUsersResponse{
		Users:      users,
		Pagination: pageRes,
	}, nil
}

func (q Querier) FastClientUser(c context.Context, req *types.QueryFastClientUserRequest) (*types.QueryFastClientUserResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	pubKeyBytes, err := hex.DecodeString(req.FastClientPubKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", req.FastClientPubKey)
	}

	fastClient, err := q.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no fast client registered for %s", req.FastClientPubKey)
		}

		return nil, err
	}

	user, err := q.GetFastUser(ctx, fastClient.Id, req.UserId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no user registered for %s", req.UserId)
		}

		return nil, err
	}

	return &types.QueryFastClientUserResponse{User: user}, nil
}

func (q Querier) FastClientEligibility(c context.Context, req *types.QueryFastClientEligibilityRequest) (*types.QueryFastClientEligibilityResponse, error) {
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

	sigPubKey, err := crypto.SigToPub(payloadHash, signatureBytes)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid signature: %s", signature)
	}

	compressedPubKey := crypto.CompressPubkey(sigPubKey)

	fastClient, err := q.GetFastClient(ctx, compressedPubKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no fast client registered for %s", hex.EncodeToString(compressedPubKey))
		}

		return nil, err
	}

	fastUser, err := q.GetFastUser(ctx, fastClient.Id, userID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("no user registered for %s", userID)
		}

		return nil, err
	}

	// The querying service is responsible for determining whether they want to handle the request or not,
	// this query simply returns whether the signature is valid, the credits of the user, and the current block height.
	return &types.QueryFastClientEligibilityResponse{
		Eligible:    true,
		UserCredits: fastUser.Credits,
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
