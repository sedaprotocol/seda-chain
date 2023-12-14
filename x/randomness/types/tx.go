package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg MsgNewSeed) Route() string {
	return RouterKey
}

func (msg MsgNewSeed) Type() string {
	return "new-seed"
}

func (msg MsgNewSeed) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return err
	}
	return nil
}
