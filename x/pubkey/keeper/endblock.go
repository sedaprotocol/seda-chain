package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in pubkey end block", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in pubkey end block", "err", err)
		}
		err = nil
	}()

	// If secp256k1 proving scheme is already enabled, do nothing.
	isEnabled, err := k.IsProvingSchemeEnabled(ctx, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	if isEnabled {
		return
	}

	// If the sum of the voting power has reached 80%, enable secp256k1
	// proving scheme.
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	var powerSum uint64
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		_, err := k.GetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return false
			} else {
				panic(err)
			}
		} else {
			//nolint:gosec // G115: We shouldn't get negative power anyway.
			powerSum += uint64(power)
		}
		return false
	})
	if err != nil {
		return err
	}

	gotPower := powerSum * 100
	//nolint:gosec // G115: We shouldn't get negative power anyway.
	requiredPower := uint64(totalPower.Int64()*100*4/5 + 1)
	if gotPower >= requiredPower {
		err = k.EnableProvingScheme(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
		// TODO: Jail validators (active and inactive) without required
		// public keys.
	}
	k.Logger(ctx).Info("checked status of secp256k1 proving scheme", "required", requiredPower, "got", gotPower)
	return
}
