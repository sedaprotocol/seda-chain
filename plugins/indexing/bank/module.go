package bank

import (
	"bytes"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	types "github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const StoreKey = banktypes.StoreKey

func ExtractUpdate(_ codec.Codec, change *storetypes.StoreKVPair) (*types.Message, error) {
	if keyBytes, found := bytes.CutPrefix(change.Key, banktypes.SupplyKey); found {
		_, key, err := collections.StringKey.Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		var amount math.Int
		err = amount.Unmarshal(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		}{
			Denom:  key,
			Amount: amount.String(),
		}

		return types.NewMessage("supply", data), nil
	} else if keyBytes, found := bytes.CutPrefix(change.Key, banktypes.BalancesPrefix); found {
		_, key, err := collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey).Decode(keyBytes)
		if err != nil {
			return nil, err
		}

		var balance math.Int
		err = balance.Unmarshal(change.Value)
		if err != nil {
			return nil, err
		}

		data := struct {
			Address string `json:"address"`
			Balance string `json:"balance"`
			Denom   string `json:"denom"`
		}{
			Address: key.K1().String(),
			Balance: balance.String(),
			Denom:   key.K2(),
		}

		return types.NewMessage("account-balance", data), nil
	}

	// TODO(#217) Log warning ("unable to process change %v", change)
	return nil, nil
}
