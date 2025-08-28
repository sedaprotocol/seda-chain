package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	err := k.setCurrentSophonID(ctx, data.StartingSophonId)
	if err != nil {
		panic(err)
	}

	for _, sophonInfo := range data.SophonInfos {
		if err := k.SetSophonInfo(ctx, sophonInfo.PublicKey, sophonInfo); err != nil {
			panic(err)
		}
	}

	for _, sophonUser := range data.SophonUsers {
		if err := k.SetSophonUser(ctx, sophonUser.SophonId, sophonUser.User.UserId, *sophonUser.User); err != nil {
			panic(err)
		}
	}

	for _, transfer := range data.Transfers {
		address, err := sdk.AccAddressFromBech32(transfer.NewOwnerAddress)
		if err != nil {
			panic(err)
		}

		if err := k.SetSophonTransfer(ctx, transfer.SophonId, address.Bytes()); err != nil {
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

	gs.StartingSophonId, err = k.sophonID.Peek(ctx)
	if err != nil {
		panic(err)
	}

	err = k.sophonInfo.Walk(ctx, nil, func(_ []byte, value types.SophonInfo) (stop bool, err error) {
		gs.SophonInfos = append(gs.SophonInfos, value)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	err = k.sophonUser.Walk(ctx, nil, func(key collections.Pair[uint64, string], value types.SophonUser) (stop bool, err error) {
		gs.SophonUsers = append(gs.SophonUsers, types.UserWithSophonId{
			SophonId: key.K1(),
			User:     &value,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	err = k.sophonTransfer.Walk(ctx, nil, func(key uint64, value []byte) (stop bool, err error) {
		gs.Transfers = append(gs.Transfers, types.SophonTransferOwnership{
			SophonId:        key,
			NewOwnerAddress: sdk.AccAddress(value).String(),
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return gs
}
