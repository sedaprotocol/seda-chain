package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
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
			err = k.SetValidatorKeyAtIndex(ctx, valAddr, sedatypes.SEDAKeyIndex(pk.Index), pk.PubKey)
			if err != nil {
				panic(err)
			}
		}
	}
	for _, scheme := range data.ProvingSchemes {
		err := k.SetProvingScheme(ctx, scheme)
		if err != nil {
			panic(err)
		}
	}
	err := k.params.Set(ctx, data.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState
	var err error
	gs.ValidatorPubKeys, err = k.GetAllValidatorPubKeys(ctx)
	if err != nil {
		panic(err)
	}
	gs.ProvingSchemes, err = k.GetAllProvingSchemes(ctx)
	if err != nil {
		panic(err)
	}
	gs.Params, err = k.params.Get(ctx)
	if err != nil {
		panic(err)
	}
	return gs
}
