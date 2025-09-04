package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(params Params, fastClients []FastClient, fastUsers []UserWithFastClientId, transfers []FastClientTransferOwnership) GenesisState {
	return GenesisState{
		Params:      params,
		FastClients: fastClients,
		FastUsers:   fastUsers,
		Transfers:   transfers,
	}
}

func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), []FastClient{}, []UserWithFastClientId{}, []FastClientTransferOwnership{})
	return &state
}

func ValidateGenesis(data GenesisState) error {
	fastClients := make(map[uint64]any)

	for _, fastClient := range data.FastClients {
		if err := fastClient.ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "FastClient validation failed")
		}

		fastClients[fastClient.Id] = fastClient
	}

	for _, fastUser := range data.FastUsers {
		if err := fastUser.User.ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "FastUser validation failed")
		}

		if _, ok := fastClients[fastUser.FastClientId]; !ok {
			return fmt.Errorf("fast client (%d) not found for fast user %x", fastUser.FastClientId, fastUser.User.UserId)
		}
	}

	for _, transfer := range data.Transfers {
		_, err := sdk.AccAddressFromBech32(transfer.NewOwnerAddress)
		if err != nil {
			return errorsmod.Wrap(err, "invalid address in transfers")
		}

		if _, ok := fastClients[transfer.FastClientId]; !ok {
			return fmt.Errorf("fast client (%d) not found for transfer to %s", transfer.FastClientId, transfer.NewOwnerAddress)
		}
	}

	return data.Params.ValidateBasic()
}
