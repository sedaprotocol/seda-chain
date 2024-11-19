package batching

import (
	"bytes"
	"encoding/hex"
	"strconv"

	// "cosmossdk.io/collections"
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

const StoreKey = batchingtypes.StoreKey

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if change.Delete {
		logger.Trace("skipping delete", "change", change)
		return nil, nil
	}

	if _, found := bytes.CutPrefix(change.Key, batchingtypes.DataResultsPrefix); found {
		val, err := codec.CollValue[batchingtypes.DataResult](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ID             string `json:"result_id"`
			DrID           string `json:"dr_id"`
			Version        string `json:"version"`
			BlockHeight    string `json:"block_height"`
			ExitCode       uint32 `json:"exit_code"`
			GasUsed        string `json:"gas_used"`
			Result         []byte `json:"result"`
			PaybackAddress string `json:"payback_address"`
			SedaPayload    string `json:"seda_payload"`
			Consensus      bool   `json:"consensus"`
		}{
			ID:             val.Id,
			DrID:           val.DrId,
			Version:        val.Version,
			BlockHeight:    strconv.FormatUint(val.BlockHeight, 10),
			ExitCode:       val.ExitCode,
			GasUsed:        strconv.FormatUint(val.GasUsed, 10),
			Result:         val.Result,
			PaybackAddress: val.PaybackAddress,
			SedaPayload:    val.SedaPayload,
			Consensus:      val.Consensus,
		}

		return types.NewMessage("data-result", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, batchingtypes.BatchesKeyPrefix); found {
		val, err := codec.CollValue[batchingtypes.Batch](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			BatchNumber     string `json:"batch_number"`
			BlockHeight     string `json:"block_height"`
			DataResultRoot  string `json:"data_result_root"`
			ValidatorRoot   string `json:"validator_root"`
			BatchID         string `json:"batch_id"`
			ProvingMetadata string `json:"proving_metadata"`
		}{
			BatchNumber:     strconv.FormatUint(val.BatchNumber, 10),
			BlockHeight:     strconv.FormatInt(val.BlockHeight, 10),
			DataResultRoot:  val.DataResultRoot,
			ValidatorRoot:   val.ValidatorRoot,
			BatchID:         hex.EncodeToString(val.BatchId),
			ProvingMetadata: hex.EncodeToString(val.ProvingMetadata),
		}

		return types.NewMessage("batch", data, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, batchingtypes.BatchAssignmentsPrefix); found {
		_, key, err := collections.StringKey.Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		val, err := collections.Uint64Value.Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			DrID        string `json:"dr_id"`
			BatchNumber string `json:"batch_number"`
		}{
			DrID:        key,
			BatchNumber: strconv.FormatUint(val, 10),
		}

		return types.NewMessage("dr-batch-assignment", data, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, batchingtypes.BatchSignaturesKeyPrefix); found {
		_, key, err := collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey).Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		val, err := codec.CollValue[batchingtypes.BatchSignatures](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			BatchNumber      string `json:"batch_number"`
			ValidatorAddress string `json:"validator_address"`
			Signatures       string `json:"signatures"`
		}{
			BatchNumber:      strconv.FormatUint(key.K1(), 10),
			ValidatorAddress: val.ValidatorAddr,
			Signatures:       hex.EncodeToString(val.Signatures),
		}

		return types.NewMessage("batch-signatures", data, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, batchingtypes.TreeEntriesKeyPrefix); found {
		_, key, err := collections.Uint64Key.Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		val, err := codec.CollValue[batchingtypes.TreeEntries](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		validatorTreeEntries := make([]string, len(val.ValidatorEntries))
		for i, entry := range val.ValidatorEntries {
			validatorTreeEntries[i] = hex.EncodeToString(entry)
		}

		data := struct {
			BatchNumber          string   `json:"batch_number"`
			ValidatorTreeEntries []string `json:"validator_tree_entries"`
		}{
			BatchNumber:          strconv.FormatUint(key, 10),
			ValidatorTreeEntries: validatorTreeEntries,
		}

		return types.NewMessage("batch-validator-entries", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, batchingtypes.ParamsKey); found {
		val, err := codec.CollValue[batchingtypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string               `json:"moduleName"`
			Params     batchingtypes.Params `json:"params"`
		}{
			ModuleName: "batching",
			Params:     val,
		}

		return types.NewMessage("module-params", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
