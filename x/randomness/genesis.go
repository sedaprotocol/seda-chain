package randomness

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/randomness/keeper"
	"github.com/sedaprotocol/seda-chain/x/randomness/types"
)

// InitGenesis puts data from genesis state into store.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	k.SetSeed(ctx, data.Seed)
}

// ExportGenesis extracts data from store to genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	seed, err := k.GetSeed(ctx)
	if err != nil {
		return types.GenesisState{}
	}

	return types.GenesisState{
		Seed: seed,
	}
}
