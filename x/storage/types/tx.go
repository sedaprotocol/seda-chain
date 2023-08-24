package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg MsgStore) Route() string {
	return RouterKey
}

func (msg MsgStore) Type() string {
	return "store"
}

func (msg MsgStore) ValidateBasic() error {
	// TO-DO
	return nil
}

func (msg MsgStore) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStore) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
