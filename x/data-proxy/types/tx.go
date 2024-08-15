package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgRegisterDataProxy) Validate() error {
	if m.PayoutAddress == "" {
		return ErrEmptyValue.Wrap("empty payout address")
	}
	if _, err := sdk.AccAddressFromBech32(m.PayoutAddress); err != nil {
		return err
	}
	if m.Fee == nil {
		return ErrEmptyValue.Wrap("empty fee")
	}
	if m.PubKey == "" {
		return ErrEmptyValue.Wrap("empty public key")
	}
	if m.Signature == "" {
		return ErrEmptyValue.Wrap("empty signature")
	}
	return nil
}

func (m *MsgEditDataProxy) Validate() error {
	// TODO
	return nil
}

func (m *MsgUpdateParams) Validate() error {
	// TODO
	return nil
}
