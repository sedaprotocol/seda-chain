package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func NewMsgClawback(funder, vestingAccount sdk.AccAddress) *MsgClawback {
	return &MsgClawback{
		FunderAddress:  funder.String(),
		AccountAddress: vestingAccount.String(),
	}
}
