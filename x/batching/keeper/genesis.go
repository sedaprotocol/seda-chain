package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// InitGenesis puts all data from genesis state into store.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	err := k.setCurrentBatchNum(ctx, data.CurrentBatchNumber)
	if err != nil {
		panic(err)
	}
	if err := k.firstBatchNumber.Set(ctx, data.FirstBatchNumber); err != nil {
		panic(err)
	}
	for _, batch := range data.Batches {
		err := k.setBatch(ctx, batch)
		if err != nil {
			panic(err)
		}
	}
	for _, data := range data.BatchData {
		err := k.setDataResultTreeEntry(ctx, data.BatchNumber, data.DataResultEntries)
		if err != nil {
			panic(err)
		}
		for _, valEntry := range data.ValidatorEntries {
			err := k.setValidatorTreeEntry(ctx, data.BatchNumber, valEntry)
			if err != nil {
				panic(err)
			}
		}
		for _, sig := range data.BatchSignatures {
			err := k.SetBatchSigSecp256k1(ctx, data.BatchNumber, sig.ValidatorAddress, sig.Secp256K1Signature)
			if err != nil {
				panic(err)
			}
		}
	}
	for _, dr := range data.DataResults {
		err := k.dataResults.Set(ctx, collections.Join3(dr.Batched, dr.DataResult.DrId, dr.DataResult.DrBlockHeight), dr.DataResult)
		if err != nil {
			panic(err)
		}
	}
	for _, batchAssignment := range data.BatchAssignments {
		err := k.SetBatchAssignment(ctx, batchAssignment.DataRequestId, batchAssignment.DataRequestHeight, batchAssignment.BatchNumber)
		if err != nil {
			panic(err)
		}
	}
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	dataResults, err := k.getAllGenesisDataResults(ctx)
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
	firstBatchNumber, err := k.firstBatchNumber.Get(ctx)
	if err != nil {
		panic(err)
	}
	batches, err := k.GetAllBatches(ctx)
	if err != nil {
		panic(err)
	}
	batchData := make([]types.BatchData, len(batches))
	for i, batch := range batches {
		data, err := k.GetBatchData(ctx, batch.BatchNumber)
		if err != nil {
			panic(err)
		}
		batchData[i] = data
	}
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(curBatchNum, firstBatchNumber, batches, batchData, dataResults, batchAssignments, params)
}
