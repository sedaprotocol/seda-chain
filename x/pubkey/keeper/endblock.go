package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
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

	scheme, err := k.GetProvingScheme(ctx, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	if scheme.ActivationHeight != types.DefaultActivationHeight && ctx.BlockHeight() >= scheme.ActivationHeight {
		scheme.IsActivated = true
		scheme.ActivationHeight = types.DefaultActivationHeight
		err = k.SetProvingScheme(ctx, scheme)
		if err != nil {
			return err
		}

		// TODO: Jail validators (active and inactive) without required
		// public keys.
	}
	if scheme.IsActivated {
		return
	}
	activationInProgress := scheme.ActivationHeight != types.DefaultActivationHeight

	met, err := k.CheckKeyRegistrationRate(ctx, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	if (activationInProgress && !met) || (!activationInProgress && met) {
		err = k.StartProvingSchemeActivation(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
	}
	return
}

// CheckKeyRegistrationRate checks if the current registration rate of
// public keys of the given key scheme surpasses the threshold.
func (k Keeper) CheckKeyRegistrationRate(ctx sdk.Context, index utils.SEDAKeyIndex) (bool, error) {
	// If the sum of the voting power has reached 80%, enable secp256k1
	// proving scheme.
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return false, err
	}

	var powerSum uint64
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		_, err := k.GetValidatorKeyAtIndex(ctx, valAddr, index)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return false
			}
			panic(err)
		}
		//nolint:gosec // G115: We shouldn't get negative power anyway.
		powerSum += uint64(power)
		return false
	})
	if err != nil {
		return false, err
	}

	//nolint:gosec // G115: We shouldn't get negative power anyway.
	requiredPower := uint64(totalPower.Int64()*100*4/5 + 1)
	gotPower := powerSum * 100

	k.Logger(ctx).Info("checked status of secp256k1 proving scheme", "required", requiredPower, "got", gotPower)

	if gotPower >= requiredPower {
		return true, nil
	}
	return false, nil
}
