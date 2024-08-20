package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreateVestingAccount{}
	_ sdk.Msg = &MsgClawback{}
)

// NewMsgCreateVestingAccount returns a reference to a new MsgCreateVestingAccount.
func NewMsgCreateVestingAccount(fromAddr, toAddr sdk.AccAddress, amount sdk.Coins, endTime int64, disableClawback bool) *MsgCreateVestingAccount {
	return &MsgCreateVestingAccount{
		FromAddress:     fromAddr.String(),
		ToAddress:       toAddr.String(),
		Amount:          amount,
		EndTime:         endTime,
		DisableClawback: disableClawback,
	}
}

func (m *MsgCreateVestingAccount) ValidateBasic() error {
	if err := validateAmount(m.Amount); err != nil {
		return err
	}
	if m.EndTime <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid end time")
	}
	return nil
}

func NewMsgClawback(funder, vestingAccount sdk.AccAddress) *MsgClawback {
	return &MsgClawback{
		FunderAddress:  funder.String(),
		AccountAddress: vestingAccount.String(),
	}
}

func validateAmount(amount sdk.Coins) error {
	if !amount.IsValid() || !amount.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(amount.String())
	}
	return nil
}
