package batching

import (
	"bytes"

	// "cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
)

const StoreKey = batchingtypes.StoreKey

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if _, found := bytes.CutPrefix(change.Key, batchingtypes.DataResultsPrefix); found {
		if change.Delete {
			logger.Trace("skipping delete", "change", change)
			return nil, nil
		}

		val, err := codec.CollValue[batchingtypes.DataResult](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ID             string `json:"result_id"`
			DrID           string `json:"dr_id"`
			Version        string `json:"version"`
			BlockHeight    uint64 `json:"block_height"`
			ExitCode       uint32 `json:"exit_code"`
			GasUsed        uint64 `json:"gas_used"`
			Result         []byte `json:"result"`
			PaybackAddress string `json:"payback_address"`
			SedaPayload    string `json:"seda_payload"`
			Consensus      bool   `json:"consensus"`
		}{
			ID:             val.Id,
			DrID:           val.DrId,
			Version:        val.Version,
			BlockHeight:    val.BlockHeight,
			ExitCode:       val.ExitCode,
			GasUsed:        val.GasUsed,
			Result:         val.Result,
			PaybackAddress: val.PaybackAddress,
			SedaPayload:    val.SedaPayload,
			Consensus:      val.Consensus,
		}

		return types.NewMessage("data-result", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
