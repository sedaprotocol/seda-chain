package keeper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

func (k Keeper) StreamImportGenesis(ctx sdk.Context, cdc codec.JSONCodec) {
	ctx.Logger().Info("stream importing batching genesis")

	f, err := os.Open("batching_export_small.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(bufio.NewReader(f))

	// Expect '{'
	tok, err := dec.Token()
	if err != nil {
		panic(err)
	}
	if tok != json.Delim('{') {
		panic(fmt.Errorf("expected start of object, got %v", tok))
	}

	// Iterate through top-level keys
	var batchCount, batchDataCount, dataResultCount, batchAssignmentCount int
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			panic(err)
		}
		key := t.(string)

		switch key {
		case "current_batch_number":
			var v uint64
			if err := dec.Decode(&v); err != nil {
				panic(err)
			}
			if err := k.setCurrentBatchNum(ctx, v); err != nil {
				panic(err)
			}

		case "batches":
			if err := streamArray(dec, func(raw json.RawMessage) error {
				var batch types.Batch
				if err := cdc.UnmarshalJSON(raw, &batch); err != nil {
					return err
				}
				err := k.setBatch(ctx, batch)
				if err != nil {
					return err
				}
				batchCount++
				if batchCount%1000 == 0 {
					ctx.Logger().Info(fmt.Sprintf("imported %d-th batch", batchCount))
				}
				return nil
			}); err != nil {
				panic(fmt.Errorf("error importing batches: %w", err))
			}

		case "batch_data":
			if err := streamArray(dec, func(raw json.RawMessage) error {
				var data types.BatchData
				if err := cdc.UnmarshalJSON(raw, &data); err != nil {
					return err
				}
				if err := k.setDataResultTreeEntry(ctx, data.BatchNumber, data.DataResultEntries); err != nil {
					return err
				}
				for _, valEntry := range data.ValidatorEntries {
					if err := k.setValidatorTreeEntry(ctx, data.BatchNumber, valEntry); err != nil {
						return err
					}
				}
				for _, sig := range data.BatchSignatures {
					if err := k.SetBatchSigSecp256k1(ctx, data.BatchNumber, sig.ValidatorAddress, sig.Secp256K1Signature); err != nil {
						return err
					}
				}
				batchDataCount++
				if batchDataCount%1000 == 0 {
					ctx.Logger().Info(fmt.Sprintf("imported %d-th batch_data", batchDataCount))
				}
				return nil
			}); err != nil {
				panic(fmt.Errorf("error importing batch_data: %w", err))
			}

		case "data_results":
			if err := streamArray(dec, func(raw json.RawMessage) error {
				var dataResult types.GenesisDataResult
				if err := cdc.UnmarshalJSON(raw, &dataResult); err != nil {
					return err
				}
				err = k.dataResults.Set(ctx, collections.Join3(false, dataResult.DataResult.DrId, dataResult.DataResult.DrBlockHeight), dataResult.DataResult)
				if err != nil {
					return err
				}
				dataResultCount++
				if dataResultCount%1000 == 0 {
					ctx.Logger().Info(fmt.Sprintf("imported %d-th data_result", dataResultCount))
				}
				return nil
			}); err != nil {
				panic(fmt.Errorf("error importing data_results: %w", err))
			}

		// -------------------------------------------------------
		// batch_assignments: [ {...}, {...}, ... ]
		// -------------------------------------------------------
		case "batch_assignments":
			if err := streamArray(dec, func(raw json.RawMessage) error {
				var ba types.BatchAssignment
				if err := cdc.UnmarshalJSON(raw, &ba); err != nil {
					return err
				}
				err := k.SetBatchAssignment(ctx, ba.DataRequestId, ba.DataRequestHeight, ba.BatchNumber)
				if err != nil {
					return err
				}
				batchAssignmentCount++
				if batchAssignmentCount%1000 == 0 {
					ctx.Logger().Info(fmt.Sprintf("imported %d-th batch_assignment", batchAssignmentCount))
				}
				return nil
			}); err != nil {
				panic(fmt.Errorf("error importing batch_assignments: %w", err))
			}

		default:
			panic(fmt.Errorf("unexpected key in export: %s", key))
		}
	}

	// Expect '}'
	if _, err := dec.Token(); err != nil {
		panic(err)
	}

	err = k.SetParams(ctx, types.DefaultParams())
	if err != nil {
		panic(err)
	}

	ctx.Logger().Info(fmt.Sprintf("stream imported %d batches, %d batch_data, %d data_results, %d batch_assignments", batchCount, batchDataCount, dataResultCount, batchAssignmentCount))
}

// streamArray reads a JSON array as a sequence of RawMessage elements.
func streamArray(dec *json.Decoder, handle func(json.RawMessage) error) error {
	// Expect '['
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('[') {
		return fmt.Errorf("expected [, got %v", tok)
	}

	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return err
		}
		if err := handle(raw); err != nil {
			return err
		}
	}

	// Expect ']'
	tok, err = dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim(']') {
		return fmt.Errorf("expected ], got %v", tok)
	}
	return nil
}

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
		if err := k.dataResults.Set(ctx, collections.Join3(dataResult.Batched, dataResult.DataResult.DrId, dataResult.DataResult.DrBlockHeight), dataResult.DataResult); err != nil {
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
	exportBatchCount := 100000 // batch and batch data
	exportDataResultCount := 100000
	exportBatchAssignmentCount := 100000

	var writer io.Writer
	f, err := os.Create("batching_export_small.json")
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
	rng := new(collections.Range[int64]).Descending()
	batchesIter, err := k.batches.Iterate(ctx, rng)
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
		if exportBatchCount != 0 && count == exportBatchCount {
			break
		}
	}
	writer.Write([]byte("],"))

	// Export batch_data
	writer.Write([]byte(`"batch_data":[`))

	first = true
	count = 0
	for batchNum := curBatchNum - 1; batchNum >= 0; batchNum-- {
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
		if exportDataResultCount != 0 && count == exportDataResultCount {
			break
		}
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

		dataResult, err := dataResultsIter.KeyValue()
		if err != nil {
			panic(err)
		}
		dataResultJSON, err := cdc.MarshalJSON(&types.GenesisDataResult{
			Batched:    dataResult.Key.K1(),
			DataResult: dataResult.Value,
		})
		if err != nil {
			panic(err)
		}
		writer.Write(dataResultJSON)

		count++
		if count%1000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("exported %d-th data_result", count))
		}
		if exportBatchAssignmentCount != 0 && count == exportBatchAssignmentCount {
			break
		}
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
		if exportBatchAssignmentCount != 0 && count == exportBatchAssignmentCount {
			break
		}
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
