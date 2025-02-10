package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
)

var (
	_ sdk.Msg = &MsgStoreOracleProgram{}
	_ sdk.Msg = &MsgInstantiateCoreContract{}
)

func (msg *MsgStoreOracleProgram) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if msg.StorageFee.AmountOf(appparams.DefaultBondDenom).IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("storage fee must be greater than 0aseda")
	}

	return validateWasmSize(msg.Wasm)
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
	return wasmtypes.ValidateSalt(msg.Salt)
}
