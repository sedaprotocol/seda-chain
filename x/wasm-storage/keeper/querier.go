package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerierImpl(keeper Keeper) types.QueryServer {
	return &Querier{
		keeper,
	}
}

func (q Querier) OracleProgram(c context.Context, req *types.QueryOracleProgramRequest) (*types.QueryOracleProgramResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	program, err := q.GetOracleProgram(ctx, req.Hash)
	if err != nil {
		return nil, err
	}
	return &types.QueryOracleProgramResponse{
		OracleProgram: &program,
	}, nil
}

func (q Querier) OraclePrograms(c context.Context, req *types.QueryOracleProgramsRequest) (*types.QueryOracleProgramsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	results, pageRes, err := query.CollectionPaginate(
		ctx, q.Keeper.OracleProgram, req.Pagination,
		func(_ []byte, v types.OracleProgram) (string, error) {
			return hex.EncodeToString(v.Hash), nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryOracleProgramsResponse{
		List:       results,
		Pagination: pageRes,
	}, nil
}

func (q Querier) CoreContractRegistry(c context.Context, _ *types.QueryCoreContractRegistryRequest) (*types.QueryCoreContractRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	coreAddress, err := q.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryCoreContractRegistryResponse{
		Address: coreAddress.String(),
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
