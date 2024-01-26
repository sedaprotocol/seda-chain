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
	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return nil, err
	}
	return &types.QueryDataRequestWasmResponse{
		Wasm: q.GetDataRequestWasm(ctx, hash),
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
	return &types.QueryOverlayWasmResponse{
		Wasm: q.GetOverlayWasm(ctx, hash),
	}, nil
}

func (q Querier) OverlayWasms(c context.Context, _ *types.QueryOverlayWasmsRequest) (*types.QueryOverlayWasmsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryOverlayWasmsResponse{
		HashTypePairs: q.ListOverlayWasms(ctx),
	}, nil
}

func (q Querier) ProxyContractRegistry(c context.Context, _ *types.QueryProxyContractRegistryRequest) (*types.QueryProxyContractRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryProxyContractRegistryResponse{
		Address: q.GetProxyContractRegistry(ctx).String(),
	}, nil
}
