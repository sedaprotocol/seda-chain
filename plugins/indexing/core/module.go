package core

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
)

const StoreKey = coretypes.StoreKey

type Params coretypes.Params

func (p Params) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if _, found := bytes.CutPrefix(change.Key, coretypes.ParamsKey); found {
		val, err := codec.CollValue[coretypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string `json:"moduleName"`
			Params     Params `json:"params"`
		}{
			ModuleName: "core",
			Params:     Params(val),
		}

		return types.NewMessage("module-params", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
