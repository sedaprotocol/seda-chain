package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewGenesisState(params Params, sophonInfos []SophonInfo, sophonUsers []UserWithSophonId, transfers []SophonTransferOwnership) GenesisState {
	return GenesisState{
		Params:      params,
		SophonInfos: sophonInfos,
		SophonUsers: sophonUsers,
		Transfers:   transfers,
	}
}

func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), []SophonInfo{}, []UserWithSophonId{}, []SophonTransferOwnership{})
	return &state
}

func ValidateGenesis(data GenesisState) error {
	sophonInfos := make(map[uint64]any)

	for _, sophonInfo := range data.SophonInfos {
		if err := sophonInfo.ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "SophonInfo validation failed")
		}

		sophonInfos[sophonInfo.Id] = sophonInfo
	}

	for _, sophonUser := range data.SophonUsers {
		if err := sophonUser.User.ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "SophonUser validation failed")
		}

		if _, ok := sophonInfos[sophonUser.SophonId]; !ok {
			return fmt.Errorf("sophon info (%d) not found for sophon user %x", sophonUser.SophonId, sophonUser.User.UserId)
		}
	}

	for _, transfer := range data.Transfers {
		_, err := sdk.AccAddressFromBech32(transfer.NewOwnerAddress)
		if err != nil {
			return errorsmod.Wrap(err, "invalid address in transfers")
		}

		if _, ok := sophonInfos[transfer.SophonId]; !ok {
			return fmt.Errorf("sophon info (%d) not found for transfer to %s", transfer.SophonId, transfer.NewOwnerAddress)
		}
	}

	return data.Params.ValidateBasic()
}
