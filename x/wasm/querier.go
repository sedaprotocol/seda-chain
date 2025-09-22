package wasm

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Querier creates a new grpc querier instance
func Querier(k *Keeper) *GrpcQuerier {
	return NewGrpcQuerier(k)
}

var _ types.QueryServer = &GrpcQuerier{}

type GrpcQuerier struct {
	*keeper.GrpcQuerier
	Keeper *Keeper
	cdc    codec.Codec
}

// NewGrpcQuerier constructor
func NewGrpcQuerier(k *Keeper) *GrpcQuerier {
	return &GrpcQuerier{
		GrpcQuerier: keeper.NewGrpcQuerier(k.cdc, k.storeService, k, k.queryGasLimit),
		Keeper:      k,
		cdc:         k.cdc,
	}
}

func (q GrpcQuerier) SmartContractState(c context.Context, req *types.QuerySmartContractStateRequest) (rsp *types.QuerySmartContractStateResponse, err error) {
	ctx := sdk.UnwrapSDKContext(c)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.QueryData.ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query data")
	}
	contractAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	coreContractAddr, err := q.Keeper.WasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}

	if contractAddr.String() != coreContractAddr.String() {
		return q.GrpcQuerier.SmartContractState(c, req)
	}

	var query *CoreContractQuery
	err = json.Unmarshal(req.QueryData, &query)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to unmarshal core contract query: %v", err)
	}

	// Encode and dispatch.
	var encodedQuery []byte
	var path string
	switch {
	case query.GetDataRequest != nil:
		encodedQuery, path, err = query.GetDataRequest.ToModuleQuery()
	case query.GetDataRequestsByStatus != nil:
		encodedQuery, path, err = query.GetDataRequestsByStatus.ToModuleQuery()
	case query.GetDataRequestsStatuses != nil:
		encodedQuery, path, err = query.GetDataRequestsStatuses.ToModuleQuery()
	case query.GetStaker != nil:
		encodedQuery, path, err = query.GetStaker.ToModuleQuery()
	case query.GetStakerAndSeq != nil:
		encodedQuery, path, err = query.GetStakerAndSeq.ToModuleQuery()
	case query.GetStakingConfig != nil:
		encodedQuery, path, err = query.GetStakingConfig.ToModuleQuery()
	case query.GetDataRequestConfig != nil:
		encodedQuery, path, err = query.GetDataRequestConfig.ToModuleQuery()
	case query.GetExecutors != nil:
		encodedQuery, path, err = query.GetExecutors.ToModuleQuery()
	case query.IsExecutorEligible != nil:
		encodedQuery, path, err = query.IsExecutorEligible.ToModuleQuery()
	default:
		// TODO Do not include query data in the error message.
		return nil, fmt.Errorf("unsupported core contract query type %s", string(req.QueryData))
	}
	if err != nil {
		return nil, err
	}

	handler := q.Keeper.queryRouter.Route(path)
	if handler == nil {
		return nil, fmt.Errorf("failed to find handler for query route %s", path)
	}

	result, err := handler(ctx, &abci.RequestQuery{
		Path: path,
		Data: encodedQuery,
		// Height: req.Height,
		// Prove: req.Prove,
	})
	if err != nil {
		return nil, err
	}

	// Decode the response.
	var responseBytes []byte
	switch {
	case query.GetDataRequest != nil:
		responseBytes, err = query.GetDataRequest.FromModuleQuery(q.cdc, result.Value)
	case query.GetDataRequestsByStatus != nil:
		responseBytes, err = query.GetDataRequestsByStatus.FromModuleQuery(q.cdc, result.Value)
	case query.GetDataRequestsStatuses != nil:
		responseBytes, err = query.GetDataRequestsStatuses.FromModuleQuery(q.cdc, result.Value)
	case query.GetStaker != nil:
		responseBytes, err = query.GetStaker.FromModuleQuery(q.cdc, result.Value)
	case query.GetStakerAndSeq != nil:
		responseBytes, err = query.GetStakerAndSeq.FromModuleQuery(q.cdc, result.Value)
	case query.GetStakingConfig != nil:
		responseBytes, err = query.GetStakingConfig.FromModuleQuery(q.cdc, result.Value)
	case query.GetDataRequestConfig != nil:
		responseBytes, err = query.GetDataRequestConfig.FromModuleQuery(q.cdc, result.Value)
	case query.GetExecutors != nil:
		responseBytes, err = query.GetExecutors.FromModuleQuery(q.cdc, result.Value)
	case query.IsExecutorEligible != nil:
		responseBytes, err = query.IsExecutorEligible.FromModuleQuery(q.cdc, result.Value)
	default:
		// TODO Do not include query data in the error message.
		return nil, fmt.Errorf("unsupported core contract query type %s", string(req.QueryData))
	}
	if err != nil {
		return nil, err
	}

	return &types.QuerySmartContractStateResponse{
		Data: types.RawContractMessage(responseBytes),
	}, nil
}
