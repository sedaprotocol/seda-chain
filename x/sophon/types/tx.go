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
	return fmt.Errorf("not implemented")
}

func (m *MsgTransferOwnership) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgAcceptOwnership) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgCancelOwnershipTransfer) ValidateBasic() error {
	return fmt.Errorf("not implemented")
}

func (m *MsgAddUser) ValidateBasic() error {
	return fmt.Errorf("not implemented")
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
	return fmt.Errorf("not implemented")
}

func (m *MsgUpdateParams) ValidateBasic() error {
	return m.Params.ValidateBasic()
}
