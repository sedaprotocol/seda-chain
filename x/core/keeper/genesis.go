package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	// set owner state
	err := k.SetOwner(ctx, data.Owner)
	if err != nil {
		panic(err)
	}

	// set pending owner state
	err = k.SetPendingOwner(ctx, data.PendingOwner)
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

	err = k.dataRequestIndexing.committingCount.Set(ctx, data.CommittingCount)
	if err != nil {
		panic(err)
	}
	err = k.dataRequestIndexing.revealingCount.Set(ctx, data.RevealingCount)
	if err != nil {
		panic(err)
	}
	err = k.dataRequestIndexing.tallyingCount.Set(ctx, data.TallyingCount)
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
	gs.PendingOwner, err = k.GetPendingOwner(ctx)
	if err != nil {
		panic(err)
	}
	gs.Paused, err = k.IsPaused(ctx)
	if err != nil {
		panic(err)
	}

	gs.CommittingCount, err = k.dataRequestIndexing.committingCount.Get(ctx)
	if err != nil {
		panic(err)
	}
	gs.RevealingCount, err = k.dataRequestIndexing.revealingCount.Get(ctx)
	if err != nil {
		panic(err)
	}
	gs.TallyingCount, err = k.dataRequestIndexing.tallyingCount.Get(ctx)
	if err != nil {
		panic(err)
	}

	return gs
}
