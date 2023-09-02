package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg MsgStoreDataRequestWasm) Route() string {
	return RouterKey
}

func (msg MsgStoreDataRequestWasm) Type() string {
	return "store-data-request-wasm"
}

func (msg MsgStoreDataRequestWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	return nil
}

func (msg MsgStoreDataRequestWasm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStoreDataRequestWasm) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as ValidateBasic() rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

func (msg MsgStoreOverlayWasm) Route() string {
	return RouterKey
}

func (msg MsgStoreOverlayWasm) Type() string {
	return "store-overlay-wasm"
}

func (msg MsgStoreOverlayWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	return nil
}

func (msg MsgStoreOverlayWasm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgStoreOverlayWasm) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as ValidateBasic() rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
