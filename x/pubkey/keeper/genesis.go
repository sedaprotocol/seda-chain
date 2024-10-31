package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	for _, val := range data.ValidatorPubKeys {
		valAddr, err := k.validatorAddressCodec.StringToBytes(val.ValidatorAddr)
		if err != nil {
			panic(err)
		}
		for _, pk := range val.IndexedPubKeys {
			err = k.SetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndex(pk.Index), pk.PubKey)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState

	itr, err := k.pubKeys.Iterate(ctx, nil)
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
