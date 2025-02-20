package tally

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	tallytypes "github.com/sedaprotocol/seda-chain/x/tally/types"
)

const StoreKey = tallytypes.StoreKey

type Params tallytypes.Params

func (p Params) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if _, found := bytes.CutPrefix(change.Key, tallytypes.ParamsPrefix); found {
		val, err := codec.CollValue[tallytypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string `json:"moduleName"`
			Params     Params `json:"params"`
		}{
			ModuleName: "tally",
			Params:     Params(val),
		}

		return types.NewMessage("module-params", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
