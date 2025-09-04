package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	err := k.setCurrentFastClientID(ctx, data.StartingFastClientId)
	if err != nil {
		panic(err)
	}

	for _, fastClient := range data.FastClients {
		if err := k.SetFastClient(ctx, fastClient.PublicKey, fastClient); err != nil {
			panic(err)
		}
	}

	for _, fastClientUser := range data.FastUsers {
		if err := k.SetFastUser(ctx, fastClientUser.FastClientId, fastClientUser.User.UserId, *fastClientUser.User); err != nil {
			panic(err)
		}
	}

	for _, transfer := range data.Transfers {
		address, err := sdk.AccAddressFromBech32(transfer.NewOwnerAddress)
		if err != nil {
			panic(err)
		}

		if err := k.SetFastTransfer(ctx, transfer.FastClientId, address.Bytes()); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState

	params, err := k.params.Get(ctx)
	if err != nil {
		panic(err)
	}
	gs.Params = params

	gs.StartingFastClientId, err = k.fastClientID.Peek(ctx)
	if err != nil {
		panic(err)
	}

	err = k.fastClient.Walk(ctx, nil, func(_ []byte, value types.FastClient) (stop bool, err error) {
		gs.FastClients = append(gs.FastClients, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	err = k.fastUser.Walk(ctx, nil, func(key collections.Pair[uint64, string], value types.FastUser) (stop bool, err error) {
		gs.FastUsers = append(gs.FastUsers, types.UserWithFastClientId{
			FastClientId: key.K1(),
			User:         &value,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	err = k.fastClientTransfer.Walk(ctx, nil, func(key uint64, value []byte) (stop bool, err error) {
		gs.Transfers = append(gs.Transfers, types.FastClientTransferOwnership{
			FastClientId:    key,
			NewOwnerAddress: sdk.AccAddress(value).String(),
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return gs
}
