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

	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return err
	}

	// If the sum of the voting power has reached (2/3 + 1), enable
	// secp256k1 proving scheme.
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
			powerSum += uint64(power)
		}
		return false
	})
	if err != nil {
		return err
	}

	if requiredPower := ((totalPower.Int64() * 4) / 5) + 1; powerSum >= uint64(requiredPower) {
		err = k.EnableProvingScheme(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
		// TODO: Jail validators (active and inactive) without required
		// public keys.
	}

	return
}
