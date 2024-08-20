package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgStoreDataRequestWasm{}
	_ sdk.Msg = &MsgStoreExecutorWasm{}
	_ sdk.Msg = &MsgInstantiateCoreContract{}
)

func (msg *MsgStoreDataRequestWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if err := validateWasmSize(msg.Wasm); err != nil {
		return err
	}
	return nil
}

func (msg *MsgStoreExecutorWasm) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if err := validateWasmSize(msg.Wasm); err != nil {
		return err
	}
	return nil
}

func (msg *MsgInstantiateCoreContract) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}
	if msg.CodeID == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("code id is required")
	}
	if err := wasmtypes.ValidateLabel(msg.Label); err != nil {
		return err
	}
	if !msg.Funds.IsValid() {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid coins")
	}
	if len(msg.Admin) != 0 {
		if _, err := sdk.AccAddressFromBech32(msg.Admin); err != nil {
			return err
		}
	}
	if err := msg.Msg.ValidateBasic(); err != nil {
		return err
	}
	if err := wasmtypes.ValidateSalt(msg.Salt); err != nil {
		return err
	}
	return nil
}
