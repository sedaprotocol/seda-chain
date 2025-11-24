package keeper

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/codec"
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
	for _, data := range data.BatchData {
		if err := k.setDataResultTreeEntry(ctx, data.BatchNumber, data.DataResultEntries); err != nil {
			panic(err)
		}
		for _, valEntry := range data.ValidatorEntries {
			if err := k.setValidatorTreeEntry(ctx, data.BatchNumber, valEntry); err != nil {
				panic(err)
			}
		}
		for _, sig := range data.BatchSignatures {
			if err := k.SetBatchSigSecp256k1(ctx, data.BatchNumber, sig.ValidatorAddress, sig.Secp256K1Signature); err != nil {
				panic(err)
			}
		}
	}
	for _, dataResult := range data.DataResults {
		if err := k.dataResults.Set(ctx, collections.Join3(false, dataResult.DataResult.DrId, dataResult.DataResult.DrBlockHeight), dataResult.DataResult); err != nil {
			panic(err)
		}
	}
	for _, batchAssignment := range data.BatchAssignments {
		if err := k.SetBatchAssignment(ctx, batchAssignment.DataRequestId, batchAssignment.DataRequestHeight, batchAssignment.BatchNumber); err != nil {
			panic(err)
		}
	}
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// In keeper/genesis.go, add a new streaming export function
func (k Keeper) StreamExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) {
	// exportBatchCount := 100000 // batch and batch data
	// exportDataResultCount := 100000
	// exportBatchAssignmentCount := 100000

	var writer io.Writer
	f, err := os.Create("betching.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	writer = f

	// Export current_batch_number
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		panic(err)
	}
	writer.Write([]byte(`{"current_batch_number":`))
	writer.Write([]byte(strconv.FormatUint(curBatchNum, 10)))
	writer.Write([]byte(","))

	// Export batches
	writer.Write([]byte(`"batches":[`))
	batchesIter, err := k.batches.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer batchesIter.Close()

	first := true
	count := 0
	for ; batchesIter.Valid(); batchesIter.Next() {
		if !first {
			writer.Write([]byte(","))
		}
		first = false

		batch, err := batchesIter.Value()
		if err != nil {
			panic(err)
		}
		batchJSON, err := cdc.MarshalJSON(&batch)
		if err != nil {
			panic(err)
		}
		writer.Write(batchJSON)

		count++
		if count%1000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("exported %d-th batch", count))
		}
	}
	writer.Write([]byte("],"))

	// Export batch_data
	writer.Write([]byte(`"batch_data":[`))

	first = true
	count = 0
	for batchNum := range curBatchNum { // TODO starting batch may not be 0 in the future
		if !first {
			writer.Write([]byte(","))
		}
		first = false

		data, err := k.GetBatchData(ctx, batchNum)
		if err != nil {
			panic(err)
		}
		batchDataJSON, err := cdc.MarshalJSON(&data)
		if err != nil {
			panic(err)
		}
		writer.Write(batchDataJSON)

		count++
		if count%1000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("exported %d-th batch_data", count))
		}
		// if count == exportBatchCount {
		// 	break
		// }
	}

	writer.Write([]byte("],"))

	// Export data_results
	writer.Write([]byte(`"data_results":[`))
	dataResultsIter, err := k.dataResults.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer dataResultsIter.Close()

	first = true
	count = 0
	for ; dataResultsIter.Valid(); dataResultsIter.Next() {
		if !first {
			writer.Write([]byte(","))
		}
		first = false

		dataResult, err := dataResultsIter.Value()
		if err != nil {
			panic(err)
		}
		dataResultJSON, err := cdc.MarshalJSON(&dataResult)
		if err != nil {
			panic(err)
		}
		writer.Write(dataResultJSON)

		count++
		if count%1000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("exported %d-th data_result", count))
		}
		// if count == exportDataResultCount {
		// 	break
		// }
	}
	writer.Write([]byte("],"))

	// Export batch_assignments
	writer.Write([]byte(`"batch_assignments":[`))
	batchAssignmentsIter, err := k.batchAssignments.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer batchAssignmentsIter.Close()

	first = true
	count = 0
	for ; batchAssignmentsIter.Valid(); batchAssignmentsIter.Next() {
		if !first {
			writer.Write([]byte(","))
		}
		first = false

		kv, err := batchAssignmentsIter.KeyValue()
		if err != nil {
			panic(err)
		}
		ba := types.BatchAssignment{
			BatchNumber:       kv.Value,
			DataRequestId:     kv.Key.K1(),
			DataRequestHeight: kv.Key.K2(),
		}
		batchAssignmentJSON, err := cdc.MarshalJSON(&ba)
		if err != nil {
			panic(err)
		}

		writer.Write(batchAssignmentJSON)

		count++
		if count%1000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("exported %d-th batch_assignment", count))
		}
		// if count == exportBatchAssignmentCount {
		// 	break
		// }
	}
	writer.Write([]byte("]"))

	writer.Write([]byte("}"))
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
	return types.NewGenesisState(curBatchNum, batches, batchData, dataResults, batchAssignments, params)
}
