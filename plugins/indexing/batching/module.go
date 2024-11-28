package batching

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"time"

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
			ID             string    `json:"result_id"`
			DrID           string    `json:"dr_id"`
			DrBlockHeight  string    `json:"dr_block_height"`
			Version        string    `json:"version"`
			BlockHeight    string    `json:"block_height"`
			Timestamp      time.Time `json:"timestamp"`
			ExitCode       uint32    `json:"exit_code"`
			GasUsed        string    `json:"gas_used"`
			Result         []byte    `json:"result"`
			PaybackAddress string    `json:"payback_address"`
			SedaPayload    string    `json:"seda_payload"`
			Consensus      bool      `json:"consensus"`
		}{
			ID:            val.Id,
			DrID:          val.DrId,
			DrBlockHeight: strconv.FormatUint(val.DrBlockHeight, 10),
			//nolint:gosec // G115: When storing the timestamp we converted from int64 to uint64, so the reverse should be safe.
			Timestamp:      time.Unix(int64(val.BlockTimestamp), 0),
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
		_, key, err := collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey).Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		// The second part of the key is empty for data result entries, we index those separately.
		if len(key.K2()) == 0 {
			return nil, nil
		}

		val, err := collections.BytesValue.Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			BatchNumber        string `json:"batch_number"`
			ValidatorTreeEntry string `json:"validator_tree_entry"`
		}{
			BatchNumber:        strconv.FormatUint(key.K1(), 10),
			ValidatorTreeEntry: hex.EncodeToString(val),
		}

		return types.NewMessage("batch-validator-entry", data, ctx), nil
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
