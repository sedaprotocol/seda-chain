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
	for i := range data.Wasms {
		wasm := data.Wasms[i]
		if wasm.WasmType == types.WasmTypeDataRequest ||
			wasm.WasmType == types.WasmTypeTally {
			err := k.SetDataRequestWasm(ctx, &wasm)
			if err != nil {
				panic(err)
			}
		}
		if wasm.WasmType == types.WasmTypeDataRequestExecutor ||
			wasm.WasmType == types.WasmTypeRelayer {
			err := k.SetOverlayWasm(ctx, &wasm)
			if err != nil {
				panic(err)
			}
		}
	}
	if data.ProxyContractRegistry != "" {
		proxyAddr, err := sdk.AccAddressFromBech32(data.ProxyContractRegistry)
		if err != nil {
			panic(err)
		}
		err = k.ProxyContractRegistry.Set(ctx, proxyAddr.String())
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	wasms := k.GetAllWasms(ctx)
	proxy, err := k.ProxyContractRegistry.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.NewGenesisState(wasms, "")
		}
		panic(err)
	}
	return types.NewGenesisState(wasms, proxy)
}
