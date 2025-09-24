package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	/*
		Module parameter states
	*/
	err := k.SetParams(ctx, data.Params)
	if err != nil {
		panic(err)
	}
	err = k.SetOwner(ctx, data.Owner)
	if err != nil {
		panic(err)
	}
	err = k.SetPendingOwner(ctx, data.PendingOwner)
	if err != nil {
		panic(err)
	}
	err = k.paused.Set(ctx, data.Paused)
	if err != nil {
		panic(err)
	}

	/*
		Staking-related states
	*/
	for _, allowlist := range data.Allowlist {
		err = k.AddToAllowlist(ctx, allowlist)
		if err != nil {
			panic(err)
		}
	}

	// Initialize staker count before setting stakers.
	err = k.stakers.SetStakerCount(ctx, 0)
	if err != nil {
		panic(err)
	}
	for _, staker := range data.Stakers {
		err = k.SetStaker(ctx, staker)
		if err != nil {
			panic(err)
		}
	}

	/*
		Data request-related states
	*/
	err = k.dataRequests.InitializeCounters(ctx)
	if err != nil {
		panic(err)
	}
	for _, dataRequest := range data.DataRequests {
		err = k.StoreDataRequest(ctx, dataRequest)
		if err != nil {
			panic(err)
		}

		err = k.AddToTimeoutQueue(ctx, dataRequest.ID, dataRequest.TimeoutHeight)
		if err != nil {
			panic(err)
		}
	}

	for _, revealBody := range data.RevealBodies {
		err = k.SetRevealBody(ctx, revealBody.Executor, revealBody.RevealBody)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState
	var err error

	/*
		Module parameter states
	*/
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

	/*
		Staking-related states
	*/
	gs.Allowlist, err = k.ListAllowlist(ctx)
	if err != nil {
		panic(err)
	}

	gs.Stakers, err = k.GetAllStakers(ctx)
	if err != nil {
		panic(err)
	}

	// Sanity check
	stakerCount, err := k.stakers.GetStakerCount(ctx)
	if err != nil {
		panic(err)
	}
	if uint64(stakerCount) != uint64(len(gs.Stakers)) {
		panic("staker count mismatch")
	}

	/*
		Data request-related states
	*/
	gs.DataRequests, err = k.GetAllDataRequests(ctx)
	if err != nil {
		panic(err)
	}

	var revealBodies []types.GenesisStateRevealBody
	for _, dataRequest := range gs.DataRequests {
		reveals, _, _ := k.LoadRevealsHashSorted(ctx, dataRequest.ID, dataRequest.Reveals, nil)
		for _, reveal := range reveals {
			revealBodies = append(revealBodies, types.GenesisStateRevealBody{Executor: reveal.Executor, RevealBody: reveal.RevealBody})
		}
	}

	gs.RevealBodies = revealBodies

	return gs
}
