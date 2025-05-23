package types

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

var (
	_ vestexported.VestingAccount = (*ClawbackContinuousVestingAccount)(nil)
	_ authtypes.GenesisAccount    = (*ClawbackContinuousVestingAccount)(nil)
)

// NewClawbackContinuousVestingAccountRaw creates a new ContinuousVestingAccount
// object from BaseVestingAccount.
func NewClawbackContinuousVestingAccountRaw(bva *vestingtypes.BaseVestingAccount, startTime int64, funder string) *ClawbackContinuousVestingAccount {
	continuousVestingAcc := vestingtypes.NewContinuousVestingAccountRaw(bva, startTime)
	return &ClawbackContinuousVestingAccount{
		ContinuousVestingAccount: continuousVestingAcc,
		FunderAddress:            funder,
	}
}
