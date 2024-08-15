package keeper

import (
	"context"
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Keeper struct {
	*sdkkeeper.Keeper
}

func NewKeeper(sdkStakingKeeper *sdkkeeper.Keeper) *Keeper {
	return &Keeper{
		Keeper: sdkStakingKeeper,
	}
}

func (k *Keeper) SetHooks(sh types.StakingHooks) {
	k.Keeper.SetHooks(sh)
}

// NOTE: This code was taken from
// https://github.com/agoric-labs/cosmos-sdk/blob/f42d86980ddfc07869846c391a03622cbd7e9188/x/staking/keeper/delegation.go#L701
// with slight modifications.
//
// TransferDelegation changes the ownership of at most the desired number of shares.
// Returns the actual number of shares transferred. Will also transfer redelegation
// entries to ensure that all redelegations are matched by sufficient shares.
// Note that no tokens are transferred to or from any pool or account, since no
// delegation is actually changing state.
func (k Keeper) TransferDelegation(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantShares sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	transferred := sdkmath.LegacyZeroDec()

	// sanity checks
	if !wantShares.IsPositive() {
		return transferred, nil
	}
	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return transferred, nil
	}
	delFrom, err := k.GetDelegation(ctx, fromAddr, valAddr)
	if err != nil {
		return transferred, nil
	}

	// Check redelegation entry limits while we can still return early.
	// Assume the worst case that we need to transfer all redelegation entries
	mightExceedLimit := false
	var cbErr error
	err = k.IterateDelegatorRedelegations(ctx, fromAddr, func(toRedelegation types.Redelegation) (stop bool) {
		// There's no redelegation index by delegator and dstVal or vice-versa.
		// The minimum cardinality is to look up by delegator, so scan and skip.
		if toRedelegation.ValidatorDstAddress != valAddr.String() {
			return false
		}
		maxEntries, err := k.MaxEntries(ctx)
		if err != nil {
			cbErr = fmt.Errorf("error getting max entries: %w", err)
			return true
		}
		valSrcAddr, err := sdk.ValAddressFromBech32(toRedelegation.ValidatorSrcAddress)
		if err != nil {
			cbErr = fmt.Errorf("error getting validator source address: %w", err)
			return true
		}
		valDstAddr, err := sdk.ValAddressFromBech32(toRedelegation.ValidatorDstAddress)
		if err != nil {
			cbErr = fmt.Errorf("error getting validator destination address: %w", err)
			return true
		}
		fromRedelegation, err := k.GetRedelegation(ctx, fromAddr, valSrcAddr, valDstAddr)
		if err == nil && len(toRedelegation.Entries)+len(fromRedelegation.Entries) >= int(maxEntries) {
			mightExceedLimit = true
			return true
		}
		return false
	})
	if cbErr != nil {
		return transferred, cbErr
	}
	if err != nil {
		return transferred, err
	}
	if mightExceedLimit {
		// avoid types.ErrMaxRedelegationEntries
		return transferred, nil
	}

	// compute shares to transfer, amount left behind
	transferred = delFrom.Shares
	if transferred.GT(wantShares) {
		transferred = wantShares
	}
	remaining := delFrom.Shares.Sub(transferred)

	// Update or create the delTo object, calling appropriate hooks
	delTo, err := k.GetDelegation(ctx, toAddr, valAddr)
	if err != nil {
		if err == types.ErrNoDelegation {
			delTo = types.NewDelegation(toAddr.String(), validator.GetOperator(), sdkmath.LegacyZeroDec())
			err = k.Hooks().BeforeDelegationCreated(ctx, toAddr, valAddr)
		} else {
			return transferred, err
		}
	} else {
		err = k.Hooks().BeforeDelegationSharesModified(ctx, toAddr, valAddr)
	}
	if err != nil {
		return transferred, nil
	}

	delTo.Shares = delTo.Shares.Add(transferred)
	err = k.SetDelegation(ctx, delTo)
	if err != nil {
		return transferred, nil
	}
	err = k.Hooks().AfterDelegationModified(ctx, toAddr, valAddr)
	if err != nil {
		return transferred, nil
	}

	// Update source delegation
	if remaining.IsZero() {
		err = k.Hooks().BeforeDelegationRemoved(ctx, fromAddr, valAddr)
		if err != nil {
			return transferred, nil
		}
		err = k.RemoveDelegation(ctx, delFrom)
		if err != nil {
			return transferred, nil
		}
	} else {
		err = k.Hooks().BeforeDelegationSharesModified(ctx, fromAddr, valAddr)
		if err != nil {
			return transferred, nil
		}
		delFrom.Shares = remaining
		err = k.SetDelegation(ctx, delFrom)
		if err != nil {
			return transferred, nil
		}
		err = k.Hooks().AfterDelegationModified(ctx, fromAddr, valAddr)
		if err != nil {
			return transferred, nil
		}
	}

	// If there are not enough remaining shares to be responsible for
	// the redelegations, transfer some redelegations.
	// For instance, if the original delegation of 300 shares to validator A
	// had redelegations for 100 shares each from validators B, C, and D,
	// and if we're transferring 175 shares, then we might keep the redelegation
	// from B, transfer the one from D, and split the redelegation from C
	// keeping a liability for 25 shares and transferring one for 75 shares.
	// Of course, the redelegations themselves can have multiple entries for
	// different timestamps, so we're actually working at a finer granularity.
	redelegations, err := k.GetRedelegations(ctx, fromAddr, math.MaxUint16)
	if err != nil {
		return transferred, err
	}
	for _, redelegation := range redelegations {
		// There's no redelegation index by delegator and dstVal or vice-versa.
		// The minimum cardinality is to look up by delegator, so scan and skip.
		if redelegation.ValidatorDstAddress != valAddr.String() {
			continue
		}

		valSrcAddr, err := sdk.ValAddressFromBech32(redelegation.ValidatorSrcAddress)
		if err != nil {
			return transferred, err
		}
		valDstAddr, err := sdk.ValAddressFromBech32(redelegation.ValidatorDstAddress)
		if err != nil {
			return transferred, err
		}

		redelegationModified := false
		entriesRemaining := false
		for i := 0; i < len(redelegation.Entries); i++ {
			entry := redelegation.Entries[i]

			// Partition SharesDst between keeping and sending
			sharesToKeep := entry.SharesDst
			sharesToSend := sdkmath.LegacyZeroDec()
			if entry.SharesDst.GT(remaining) {
				sharesToKeep = remaining
				sharesToSend = entry.SharesDst.Sub(sharesToKeep)
			}
			remaining = remaining.Sub(sharesToKeep) // fewer local shares available to cover liability

			if sharesToSend.IsZero() {
				// Leave the entry here
				entriesRemaining = true
				continue
			}
			if sharesToKeep.IsZero() {
				// Transfer the whole entry, delete locally
				toRed, err := k.SetRedelegationEntry(
					ctx, toAddr, valSrcAddr, valDstAddr,
					entry.CreationHeight, entry.CompletionTime, entry.InitialBalance, sdkmath.LegacyZeroDec(), sharesToSend,
				)
				if err != nil {
					return transferred, err
				}
				err = k.InsertRedelegationQueue(ctx, toRed, entry.CompletionTime)
				if err != nil {
					return transferred, err
				}
				redelegation.RemoveEntry(int64(i))
				i--
				// okay to leave an obsolete entry in the queue for the removed entry
				redelegationModified = true
			} else {
				// Proportionally divide the entry
				fracSending := sharesToSend.Quo(entry.SharesDst)
				balanceToSend := fracSending.MulInt(entry.InitialBalance).TruncateInt()
				balanceToKeep := entry.InitialBalance.Sub(balanceToSend)
				toRed, err := k.SetRedelegationEntry(
					ctx, toAddr, valSrcAddr, valDstAddr,
					entry.CreationHeight, entry.CompletionTime, balanceToSend, sdkmath.LegacyZeroDec(), sharesToSend,
				)
				if err != nil {
					return transferred, err
				}
				err = k.InsertRedelegationQueue(ctx, toRed, entry.CompletionTime)
				if err != nil {
					return transferred, err
				}
				entry.InitialBalance = balanceToKeep
				entry.SharesDst = sharesToKeep
				redelegation.Entries[i] = entry
				// not modifying the completion time, so no need to change the queue
				redelegationModified = true
				entriesRemaining = true
			}
		}
		if redelegationModified {
			if entriesRemaining {
				err = k.SetRedelegation(ctx, redelegation)
			} else {
				err = k.RemoveRedelegation(ctx, redelegation)
			}
			if err != nil {
				return transferred, err
			}
		}
	}
	return transferred, nil
}

// NOTE: This code was taken from
// https://github.com/agoric-labs/cosmos-sdk/blob/f42d86980ddfc07869846c391a03622cbd7e9188/x/staking/keeper/delegation.go#L979
// with slight modifications.
//
// TransferUnbonding changes the ownership of UnbondingDelegation entries
// until the desired number of tokens have changed hands. Returns the actual
// number of tokens transferred.
func (k Keeper) TransferUnbonding(ctx context.Context, fromAddr, toAddr sdk.AccAddress, valAddr sdk.ValAddress, wantAmt sdkmath.Int) (sdkmath.Int, error) {
	transferred := sdkmath.ZeroInt()
	ubdFrom, err := k.GetUnbondingDelegation(ctx, fromAddr, valAddr)
	if err != nil {
		return transferred, err
	}

	ubdFromModified := false

	for i := 0; i < len(ubdFrom.Entries) && wantAmt.IsPositive(); i++ {
		entry := ubdFrom.Entries[i]
		toXfer := entry.Balance
		if toXfer.GT(wantAmt) {
			toXfer = wantAmt
		}
		if !toXfer.IsPositive() {
			continue
		}

		maxed, err := k.HasMaxUnbondingDelegationEntries(ctx, toAddr, valAddr)
		if err != nil {
			return transferred, err
		}
		if maxed {
			break
		}
		ubdTo, err := k.SetUnbondingDelegationEntry(ctx, toAddr, valAddr, entry.CreationHeight, entry.CompletionTime, toXfer)
		if err != nil {
			return transferred, err
		}
		err = k.InsertUBDQueue(ctx, ubdTo, entry.CompletionTime)
		if err != nil {
			return transferred, err
		}

		transferred = transferred.Add(toXfer)
		wantAmt = wantAmt.Sub(toXfer)

		ubdFromModified = true
		remaining := entry.Balance.Sub(toXfer)
		if remaining.IsZero() {
			ubdFrom.RemoveEntry(int64(i))
			i--
			continue
		}
		entry.Balance = remaining
		ubdFrom.Entries[i] = entry
	}

	if ubdFromModified {
		if len(ubdFrom.Entries) == 0 {
			err = k.RemoveUnbondingDelegation(ctx, ubdFrom)
		} else {
			err = k.SetUnbondingDelegation(ctx, ubdFrom)
		}
		if err != nil {
			return transferred, err
		}
	}
	return transferred, nil
}
