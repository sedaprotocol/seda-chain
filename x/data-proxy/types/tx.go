package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
)

func (m *MsgRegisterDataProxy) Validate() error {
	if m.PayoutAddress == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty payout address")
	}
	if _, err := sdk.AccAddressFromBech32(m.PayoutAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid payout address: %s", m.PayoutAddress)
	}
	if m.Fee == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("empty fee")
	}
	if m.PubKey == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}
	if m.Signature == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty signature")
	}
	if m.Fee.Denom != appparams.DefaultBondDenom {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", m.Fee.Denom, appparams.DefaultBondDenom)
	}

	return nil
}

func (m *MsgEditDataProxy) Validate() error {
	if m.PubKey == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty public key")
	}
	if m.Sender == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty sender")
	}

	hasNewPayoutAddress := m.NewPayoutAddress != DoNotModifyField
	hasNewMemo := m.NewMemo != DoNotModifyField
	hasNewFee := m.NewFee != nil

	if !hasNewPayoutAddress && !hasNewMemo && !hasNewFee {
		return ErrEmptyUpdate
	}

	if hasNewPayoutAddress {
		if _, err := sdk.AccAddressFromBech32(m.NewPayoutAddress); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid new payout address: %s", m.NewPayoutAddress)
		}
	}

	if hasNewFee {
		if m.NewFee.Denom != appparams.DefaultBondDenom {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid coin denomination: got %s, expected %s", m.NewFee.Denom, appparams.DefaultBondDenom)
		}
	}

	return nil
}

func (m *MsgTransferAdmin) Validate() error {
	if m.Sender == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty sender")
	}

	if m.NewAdminAddress == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty admin address")
	}
	if _, err := sdk.AccAddressFromBech32(m.NewAdminAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid new admin address: %s", m.NewAdminAddress)
	}

	return nil
}

func (m *MsgUpdateParams) Validate() error {
	return m.Params.Validate()
}
