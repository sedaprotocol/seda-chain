package pubkey

import (
	"bytes"
	"encoding/hex"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	"github.com/sedaprotocol/seda-chain/plugins/indexing/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

const StoreKey = pubkeytypes.StoreKey

var validatorAddressCodec = authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())

func ExtractUpdate(ctx *types.BlockContext, _ codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
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
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
