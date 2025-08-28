package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (m *MsgRegisterSophon) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority: %s", m.Authority)
	}

	if _, err := sdk.AccAddressFromBech32(m.OwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", m.OwnerAddress)
	}

	if _, err := sdk.AccAddressFromBech32(m.AdminAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", m.AdminAddress)
	}

	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", m.Address)
	}

	if len(m.PublicKey) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}

	_, err := hex.DecodeString(m.PublicKey)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.PublicKey)
	}

	if err := ValidateMemo(m.Memo); err != nil {
		return err
	}

	return nil
}

func (m *MsgEditSophon) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.OwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", m.OwnerAddress)
	}

	hasChanges := false
	if m.NewAdminAddress != DoNotModifyField {
		hasChanges = true

		if _, err := sdk.AccAddressFromBech32(m.NewAdminAddress); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", m.NewAdminAddress)
		}
	}

	if m.NewAddress != DoNotModifyField {
		hasChanges = true

		if _, err := sdk.AccAddressFromBech32(m.NewAddress); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", m.NewAddress)
		}
	}

	if m.NewPublicKey != DoNotModifyField {
		hasChanges = true

		if len(m.NewPublicKey) == 0 {
			return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
		}

		_, err := hex.DecodeString(m.NewPublicKey)
		if err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.NewPublicKey)
		}
	}

	if m.NewMemo != DoNotModifyField {
		hasChanges = true

		if err := ValidateMemo(m.NewMemo); err != nil {
			return err
		}
	}

	if !hasChanges {
		return sdkerrors.ErrInvalidRequest.Wrap("no changes provided")
	}

	return nil
}

func (m *MsgTransferOwnership) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.OwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", m.OwnerAddress)
	}

	if _, err := sdk.AccAddressFromBech32(m.NewOwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", m.NewOwnerAddress)
	}

	if m.NewOwnerAddress == m.OwnerAddress {
		return sdkerrors.ErrInvalidRequest.Wrapf("new owner address is the same as the current owner address")
	}

	if len(m.SophonPublicKey) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}

	_, err := hex.DecodeString(m.SophonPublicKey)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.SophonPublicKey)
	}

	return nil
}

func (m *MsgAcceptOwnership) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.NewOwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", m.NewOwnerAddress)
	}

	if len(m.SophonPublicKey) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}

	_, err := hex.DecodeString(m.SophonPublicKey)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.SophonPublicKey)
	}

	return nil
}

func (m *MsgCancelOwnershipTransfer) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.OwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", m.OwnerAddress)
	}

	if len(m.SophonPublicKey) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}

	_, err := hex.DecodeString(m.SophonPublicKey)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.SophonPublicKey)
	}

	return nil
}

func (m *MsgAddUser) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.AdminAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", m.AdminAddress)
	}

	if len(m.SophonPublicKey) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}

	_, err := hex.DecodeString(m.SophonPublicKey)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.SophonPublicKey)
	}

	if len(m.UserId) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("user id is empty")
	}

	if m.InitialCredits.IsNil() {
		return sdkerrors.ErrInvalidRequest.Wrap("initial credits is nil")
	}

	if m.InitialCredits.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("initial credits is negative")
	}

	return nil
}

func (m *MsgRemoveUser) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.AdminAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", m.AdminAddress)
	}

	if len(m.SophonPublicKey) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}

	_, err := hex.DecodeString(m.SophonPublicKey)
	if err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", m.SophonPublicKey)
	}

	if len(m.UserId) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("user id is empty")
	}

	return nil
}

func (m *MsgTopUpUser) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgExpireCredits) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgSettleCredits) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgSubmitReports) ValidateBasic() error {
	// TODO: Deduplicate validation logic (hex strings, credits)
	return fmt.Errorf("not implemented")
}

func (m *MsgUpdateParams) ValidateBasic() error {
	return m.Params.ValidateBasic()
}
