package wasm_storage

import (
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
			k.SetDataRequestWasm(ctx, &wasm)
		}
		if wasm.WasmType == types.WasmTypeDataRequestExecutor ||
			wasm.WasmType == types.WasmTypeRelayer {
			k.SetOverlayWasm(ctx, &wasm)
		}
	}
	if data.ProxyContractRegistry != "" {
		proxyAddr, err := sdk.AccAddressFromBech32(data.ProxyContractRegistry)
		if err != nil {
			panic(err)
		}
		k.SetProxyContractRegistry(ctx, proxyAddr)
	}
}

// ExportGenesis extracts all data from store to genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	wasms := k.GetAllWasms(ctx)
	proxy := k.GetProxyContractRegistry(ctx)
	return types.NewGenesisState(wasms, proxy.String())
}
