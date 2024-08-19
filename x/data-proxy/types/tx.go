package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgRegisterDataProxy) Validate() error {
	if m.PayoutAddress == "" {
		return ErrEmptyValue.Wrap("empty payout address")
	}
	if _, err := sdk.AccAddressFromBech32(m.PayoutAddress); err != nil {
		return ErrInvalidAddress.Wrap(err.Error())
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
	if m.PubKey == "" {
		return ErrEmptyValue.Wrap("empty public key")
	}
	if m.Sender == "" {
		return ErrEmptyValue.Wrap("empty sender")
	}

	hasNewPayoutAddress := m.NewPayoutAddress != DoNotModifyField
	hasNewMemo := m.NewMemo != DoNotModifyField
	hasNewFee := m.NewFee != nil

	if !hasNewPayoutAddress && !hasNewMemo && !hasNewFee {
		return ErrEmptyUpdate
	}

	if hasNewPayoutAddress {
		if _, err := sdk.AccAddressFromBech32(m.NewPayoutAddress); err != nil {
			return ErrInvalidAddress.Wrap(err.Error())
		}
	}

	return nil
}

func (m *MsgTransferAdmin) Validate() error {
	if m.Sender == "" {
		return ErrEmptyValue.Wrap("empty sender")
	}

	if m.NewAdminAddress == "" {
		return ErrEmptyValue.Wrap("empty admin address")
	}
	if _, err := sdk.AccAddressFromBech32(m.NewAdminAddress); err != nil {
		return ErrInvalidAddress.Wrap(err.Error())
	}

	return nil
}

func (m *MsgUpdateParams) Validate() error {
	// TODO
	return nil
}
