package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// InitGenesis puts all data from genesis state into store.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.SetCurrentBatchNum(ctx, data.CurrentBatchNumber); err != nil {
		panic(err)
	}
	for _, batch := range data.Batches {
		if err := k.SetBatch(ctx, batch); err != nil {
			panic(err)
		}
	}
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		panic(err)
	}
	batches, err := k.GetAllBatches(ctx)
	if err != nil {
		panic(err)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(curBatchNum, batches, params)
}
