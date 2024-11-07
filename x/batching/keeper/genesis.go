package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// InitGenesis puts all data from genesis state into store.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.setCurrentBatchNum(ctx, data.CurrentBatchNumber); err != nil {
		panic(err)
	}
	for _, batch := range data.Batches {
		if err := k.setBatch(ctx, batch); err != nil {
			panic(err)
		}
	}
	for _, entry := range data.TreeEntries {
		if err := k.setTreeEntry(ctx, entry); err != nil {
			panic(err)
		}
	}
	if err := k.setParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	dataResults, err := k.getAllDataResults(ctx)
	if err != nil {
		panic(err)
	}
	batchAssignments, err := k.getAllBatchAssignments(ctx)
	if err != nil {
		panic(err)
	}
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		panic(err)
	}
	batches, err := k.GetAllBatches(ctx)
	if err != nil {
		panic(err)
	}
	entries, err := k.GetAllTreeEntries(ctx)
	if err != nil {
		panic(err)
	}
	signatures, err := k.GetAllBatchSignatures(ctx)
	if err != nil {
		panic(err)
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(curBatchNum, batches, entries, dataResults, batchAssignments, signatures, params)
}
