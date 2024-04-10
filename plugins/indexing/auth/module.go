package bank

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	log "github.com/sedaprotocol/seda-chain/plugins/indexing/log"
	types "github.com/sedaprotocol/seda-chain/plugins/indexing/types"
)

const StoreKey = authtypes.StoreKey

type wrappedAccount struct {
	cdc     codec.Codec
	Account sdk.AccountI
}

func (s wrappedAccount) MarshalJSON() ([]byte, error) {
	return s.cdc.MarshalInterfaceJSON(s.Account)
}

func ExtractUpdate(ctx *types.BlockContext, cdc codec.Codec, logger *log.Logger, change *storetypes.StoreKVPair) (*types.Message, error) {
	if _, found := bytes.CutPrefix(change.Key, authtypes.AddressStoreKeyPrefix); found {
		acc, err := codec.CollInterfaceValue[sdk.AccountI](cdc).Decode(change.Value)
		if err != nil {
			return nil, err
		}

		return types.NewMessage("account", &wrappedAccount{cdc: cdc, Account: acc}, ctx), nil
	}

	logger.Trace("skipping change", "change", change)
	return nil, nil
}
