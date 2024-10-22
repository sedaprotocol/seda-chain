package oracleprogram

import (
	"bytes"
	"encoding/hex"
	"time"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	oracleprogramtypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const StoreKey = oracleprogramtypes.StoreKey

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if _, found := bytes.CutPrefix(change.Key, oracleprogramtypes.OracleProgramPrefix); found {
		val, err := codec.CollValue[oracleprogramtypes.OracleProgram](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ID               string    `json:"id"`
			Bytecode         []byte    `json:"bytecode"`
			AddedAt          time.Time `json:"addedAt"`
			ExpirationHeight int64     `json:"expirationHeight"`
		}{
			ID:               hex.EncodeToString(val.Hash),
			Bytecode:         val.Bytecode,
			AddedAt:          val.AddedAt,
			ExpirationHeight: val.ExpirationHeight,
		}

		return types.NewMessage("oracle-program", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, oracleprogramtypes.ParamsPrefix); found {
		val, err := codec.CollValue[oracleprogramtypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string                    `json:"moduleName"`
			Params     oracleprogramtypes.Params `json:"params"`
		}{
			ModuleName: "oracle-program-storage",
			Params:     val,
		}

		return types.NewMessage("module-params", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
