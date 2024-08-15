package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(_ sdk.Context, _ types.GenesisState) {
	// TODO
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(_ sdk.Context) types.GenesisState {
	var gs types.GenesisState

	// TODO

	return gs
}
