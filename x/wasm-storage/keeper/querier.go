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
		HashTypePairs: q.ListDataRequestWasms(ctx),
	}, nil
}

func (q Querier) OverlayWasm(c context.Context, req *types.QueryOverlayWasmRequest) (*types.QueryOverlayWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return nil, err
	}
	wasm, err := q.Keeper.OverlayWasm.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	return &types.QueryOverlayWasmResponse{
		Wasm: &wasm,
	}, nil
}

func (q Querier) OverlayWasms(c context.Context, _ *types.QueryOverlayWasmsRequest) (*types.QueryOverlayWasmsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryOverlayWasmsResponse{
		HashTypePairs: q.ListOverlayWasms(ctx),
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
