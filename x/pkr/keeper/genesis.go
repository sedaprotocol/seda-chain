package keeper

import (
	"bytes"

	"cosmossdk.io/collections"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	for _, val := range data.ValidatorPubKeys {
		valAddr, err := k.validatorAddressCodec.StringToBytes(val.ValidatorAddr)
		if err != nil {
			panic(err)
		}
		for _, pk := range val.IndexedPubKeys {
			pubKey, ok := pk.PubKey.GetCachedValue().(cryptotypes.PubKey)
			if !ok {
				panic("failed to unpack public key")
			}
			err = k.PubKeys.Set(ctx, collections.Join(valAddr, pk.Index), pubKey)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState

	itr, err := k.PubKeys.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer itr.Close()

	var currentVal []byte
	for ; itr.Valid(); itr.Next() {
		kv, err := itr.KeyValue()
		if err != nil {
			panic(err)
		}

		// Skip if the validator has already been processed.
		if bytes.Equal(kv.Key.K1(), currentVal) {
			continue
		}
		currentVal = kv.Key.K1()

		valAddr, err := k.validatorAddressCodec.BytesToString(kv.Key.K1())
		if err != nil {
			panic(err)
		}
		res, err := k.GetValidatorKeys(ctx, valAddr)
		if err != nil {
			panic(err)
		}

		gs.ValidatorPubKeys = append(gs.ValidatorPubKeys, res)
	}
	return gs
}
