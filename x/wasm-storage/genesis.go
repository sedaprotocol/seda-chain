package wasmstorage

import (
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// InitGenesis puts all data from genesis state into store.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}

	for i := range data.Wasms {
		wasm := data.Wasms[i]
		switch wasm.WasmType {
		case types.WasmTypeDataRequest, types.WasmTypeTally:
			if err := k.DataRequestWasm.Set(ctx, keeper.GetPrefix(wasm), wasm); err != nil {
				panic(err)
			}
		case types.WasmTypeDataRequestExecutor, types.WasmTypeRelayer:
			if err := k.OverlayWasm.Set(ctx, keeper.GetPrefix(wasm), wasm); err != nil {
				panic(err)
			}
		}
	}
	if data.ProxyContractRegistry != "" {
		proxyAddr, err := sdk.AccAddressFromBech32(data.ProxyContractRegistry)
		if err != nil {
			panic(err)
		}
		if err = k.ProxyContractRegistry.Set(ctx, proxyAddr.String()); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	wasms := k.GetAllWasms(ctx)
	proxy, err := k.ProxyContractRegistry.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.NewGenesisState(params, wasms, "")
		}
		panic(err)
	}
	return types.NewGenesisState(params, wasms, proxy)
}
