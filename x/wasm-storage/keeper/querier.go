package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (q Querier) DataRequestWasm(c context.Context, req *types.QueryDataRequestWasmRequest) (*types.QueryDataRequestWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	wasm, err := q.Keeper.GetDataRequestWasm(ctx, req.Hash)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestWasmResponse{
		Wasm: &wasm,
	}, nil
}

func (q Querier) DataRequestWasms(c context.Context, _ *types.QueryDataRequestWasmsRequest) (*types.QueryDataRequestWasmsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryDataRequestWasmsResponse{
		List: q.ListDataRequestWasms(ctx),
	}, nil
}

func (q Querier) ExecutorWasm(c context.Context, req *types.QueryExecutorWasmRequest) (*types.QueryExecutorWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return nil, err
	}
	wasm, err := q.Keeper.ExecutorWasm.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	return &types.QueryExecutorWasmResponse{
		Wasm: &wasm,
	}, nil
}

func (q Querier) ExecutorWasms(c context.Context, _ *types.QueryExecutorWasmsRequest) (*types.QueryExecutorWasmsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryExecutorWasmsResponse{
		List: q.ListExecutorWasms(ctx),
	}, nil
}

func (q Querier) CoreContractRegistry(c context.Context, _ *types.QueryCoreContractRegistryRequest) (*types.QueryCoreContractRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	coreAddress, err := q.Keeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryCoreContractRegistryResponse{
		Address: coreAddress.String(),
	}, nil
}
