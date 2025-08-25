package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
)

func NewGenesisState(params Params, sophonInfos []SophonInfo, sophonUsers []UserWithSophonId) GenesisState {
	return GenesisState{
		Params:      params,
		SophonInfos: sophonInfos,
		SophonUsers: sophonUsers,
	}
}

func DefaultGenesisState() *GenesisState {
	state := NewGenesisState(DefaultParams(), []SophonInfo{}, []UserWithSophonId{})
	return &state
}

func ValidateGenesis(data GenesisState) error {
	sophonInfos := make(map[uint64]SophonInfo)

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
		if len(transfer.NewOwnerAddress) == 0 {
			return fmt.Errorf("empty new owner address in transfers")
		}

		if _, ok := sophonInfos[transfer.SophonId]; !ok {
			return fmt.Errorf("sophon info (%d) not found for transfer to %s", transfer.SophonId, transfer.NewOwnerAddress)
		}
	}

	return data.Params.ValidateBasic()
}
