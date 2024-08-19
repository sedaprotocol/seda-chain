package types

import "cosmossdk.io/errors"

var (
	ErrNotClawbackAccount = errors.Register(ModuleName, 2, "account is not clawback continuous vesting account type")
	ErrNoVestingCoins     = errors.Register(ModuleName, 3, "vesting account does not have currently vesting coins")
	ErrNoFunderRegistered = errors.Register(ModuleName, 4, "clawback continuous vesting account does not have funder registered (clawback is disabled)")
	ErrIncompleteClawback = errors.Register(ModuleName, 5, "failed to clawback full amount")
)
