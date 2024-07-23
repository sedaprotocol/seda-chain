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
	if err := k.CoreContractRegistry.Set(ctx, data.CoreContractRegistry); err != nil {
		panic(err)
	}

	for i := range data.Wasms {
		wasm := data.Wasms[i]
		switch wasm.WasmType {
		case types.WasmTypeDataRequest:
			if err := k.DataRequestWasm.Set(ctx, wasm.Hash, wasm); err != nil {
				panic(err)
			}
		case types.WasmTypeDataRequestExecutor, types.WasmTypeRelayer:
			if err := k.OverlayWasm.Set(ctx, wasm.Hash, wasm); err != nil {
				panic(err)
			}
		default:
			panic("unexpected wasm type")
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
	core, err := k.CoreContractRegistry.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.NewGenesisState(params, wasms, "")
		}
		panic(err)
	}
	return types.NewGenesisState(params, wasms, core)
}
