package pubkey

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

const StoreKey = pubkeytypes.StoreKey

var validatorAddressCodec = authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())

type Params pubkeytypes.Params

func (p Params) MarshalJSON() ([]byte, error) {
	return types.MarshalJSJSON(p)
}

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if keyBytes, found := bytes.CutPrefix(change.Key, pubkeytypes.PubKeysPrefix); found {
		_, key, err := collections.PairKeyCodec(collections.BytesKey, collections.Uint32Key).Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		validatorAddr, err := validatorAddressCodec.BytesToString(key.K1())
		if err != nil {
			return nil, err
		}

		pubKey, err := collections.BytesValue.Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ValidatorAddress string `json:"validator_address"`
			PubKey           string `json:"pub_key"`
			Index            uint32 `json:"index"`
		}{
			ValidatorAddress: validatorAddr,
			PubKey:           hex.EncodeToString(pubKey),
			Index:            key.K2(),
		}

		return types.NewMessage("validator-pubkey", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, pubkeytypes.ProvingSchemesPrefix); found {
		val, err := codec.CollValue[pubkeytypes.ProvingScheme](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			Index            uint32 `json:"index"`
			IsActivated      bool   `json:"is_activated"`
			ActivationHeight string `json:"activation_height"`
		}{
			Index:            val.Index,
			IsActivated:      val.IsActivated,
			ActivationHeight: fmt.Sprintf("%d", val.ActivationHeight),
		}

		return types.NewMessage("proving-scheme", data, ctx), nil
	} else if _, found := bytes.CutPrefix(change.Key, pubkeytypes.ParamsPrefix); found {
		val, err := codec.CollValue[pubkeytypes.Params](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			ModuleName string `json:"moduleName"`
			Params     Params `json:"params"`
		}{
			ModuleName: "pubkey",
			Params:     Params(val),
		}

		return types.NewMessage("module-params", data, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
