package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	// set owner state
	err := k.owner.Set(ctx, data.Owner)
	if err != nil {
		panic(err)
	}

	// set paused state
	err = k.paused.Set(ctx, data.Paused)
	if err != nil {
		panic(err)
	}

	err = k.SetParams(ctx, data.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState
	var err error

	gs.Params, err = k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	gs.Owner, err = k.GetOwner(ctx)
	if err != nil {
		panic(err)
	}
	gs.Paused, err = k.IsPaused(ctx)
	if err != nil {
		panic(err)
	}
	return gs
}
