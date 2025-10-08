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

	for _, reveal := range data.Reveals {
		err = k.SetRevealBody(ctx, reveal.PublicKey, reveal.RevealBody)
		if err != nil {
			panic(err)
		}
		_, err = k.MarkAsRevealed(ctx, reveal.DrID, reveal.PublicKey)
		if err != nil {
			panic(err)
		}
	}

	for _, commit := range data.Commits {
		_, err = k.AddCommit(ctx, commit.DrID, commit.PublicKey, commit.Commit)
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

	var commitState []types.GenesisStateCommit
	var revealState []types.GenesisStateReveal
	for _, dataRequest := range gs.DataRequests {
		committers, revealers, err := k.GetCommittersAndRevealers(ctx, dataRequest.ID)
		if err != nil {
			panic(err)
		}

		for _, committer := range committers {
			commit, err := k.GetCommit(ctx, dataRequest.ID, committer)
			if err != nil {
				panic(err)
			}
			commitState = append(
				commitState,
				types.GenesisStateCommit{
					DrID:      dataRequest.ID,
					PublicKey: committer,
					Commit:    commit,
				},
			)
		}

		reveals, _, _ := k.LoadRevealsHashSorted(ctx, dataRequest.ID, revealers, nil)
		for _, reveal := range reveals {
			revealState = append(
				revealState,
				types.GenesisStateReveal{
					DrID:       dataRequest.ID,
					PublicKey:  reveal.Executor,
					RevealBody: reveal.RevealBody,
				},
			)
		}
	}

	gs.Commits = commitState
	gs.Reveals = revealState

	return gs
}
