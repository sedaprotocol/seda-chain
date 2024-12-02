package batching

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"time"

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
		_, key, err := collections.PairKeyCodec(collections.StringKey, collections.Uint64Key).Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		val, err := collections.Uint64Value.Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			DrID        string `json:"dr_id"`
			DrHeight    string `json:"dr_height"`
			BatchNumber string `json:"batch_number"`
		}{
			DrID:        key.K1(),
			DrHeight:    strconv.FormatUint(key.K2(), 10),
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
			BatchNumber        string `json:"batch_number"`
			ValidatorAddress   string `json:"validator_address"`
			Secp256K1Signature string `json:"secpk256k1_signature"`
		}{
			BatchNumber:        strconv.FormatUint(key.K1(), 10),
			ValidatorAddress:   val.ValidatorAddress.String(),
			Secp256K1Signature: hex.EncodeToString(val.Secp256K1Signature),
		}

		return types.NewMessage("batch-signatures", data, ctx), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, batchingtypes.ValidatorTreeEntriesKeyPrefix); found {
		_, key, err := collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey).Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		// The second part of the key is empty for data result entries, we index those separately.
		if len(key.K2()) == 0 {
			return nil, nil
		}

		valEntry, err := codec.CollValue[batchingtypes.ValidatorTreeEntry](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			BatchNumber        string `json:"batch_number"`
			ValidatorAddress   string `json:"validator_address"`
			VotingPowerPercent uint32 `json:"voting_power_percent"`
			EthAddress         string `json:"eth_address"`
		}{
			BatchNumber:        strconv.FormatUint(key.K1(), 10),
			ValidatorAddress:   valEntry.ValidatorAddress.String(),
			VotingPowerPercent: valEntry.VotingPowerPercent,
			EthAddress:         hex.EncodeToString(valEntry.EthAddress),
		}

		return types.NewMessage("batch-validator-entry", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
