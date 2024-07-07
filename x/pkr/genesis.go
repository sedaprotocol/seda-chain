package pkr

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/pkr/keeper"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func InitGenesis(_ sdk.Context, _ keeper.Keeper, _ types.GenesisState) {
	return
}

// ExportGenesis extracts all data from store to genesis state.
func ExportGenesis(_ sdk.Context, _ keeper.Keeper) types.GenesisState {
	return types.NewGenesisState(nil)
}
