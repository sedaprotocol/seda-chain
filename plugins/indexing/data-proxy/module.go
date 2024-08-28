package dataproxy

import (
	"bytes"
	"encoding/hex"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

const StoreKey = dataproxytypes.StoreKey

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if keyBytes, found := bytes.CutPrefix(change.Key, dataproxytypes.DataProxyConfigPrefix); found {
		_, key, err := collections.BytesKey.Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		val, err := codec.CollValue[dataproxytypes.ProxyConfig](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			PubKey string                     `json:"pubKey"`
			Config dataproxytypes.ProxyConfig `json:"config"`
		}{
			PubKey: hex.EncodeToString(key),
			Config: val,
		}

		return types.NewMessage("data-proxy", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, dataproxytypes.ParamsPrefix); found {
		val, err := codec.CollValue[dataproxytypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string                `json:"moduleName"`
			Params     dataproxytypes.Params `json:"params"`
		}{
			ModuleName: "data-proxy",
			Params:     val,
		}

		return types.NewMessage("module-params", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
