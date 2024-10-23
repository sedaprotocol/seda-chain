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

	for _, program := range data.OraclePrograms {
		if err := k.OracleProgram.Set(ctx, program.Hash, program); err != nil {
			panic(err)
		}
	}
	for _, wasm := range data.ExecutorWasms {
		if err := k.ExecutorWasm.Set(ctx, wasm.Hash, wasm); err != nil {
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
	programs := k.GetAllOraclePrograms(ctx)
	execWasms := k.GetAllExecutorWasms(ctx)
	core, err := k.GetCoreContractAddr(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.NewGenesisState(params, programs, execWasms, "")
		}
		panic(err)
	}
	return types.NewGenesisState(params, programs, execWasms, core.String())
}
