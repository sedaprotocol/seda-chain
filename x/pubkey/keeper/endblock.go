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

	// If the proving scheme is already activated, do nothing.
	if scheme.IsActivated {
		return
	}

	// Process activation in progress.
	activationInProgress := scheme.ActivationHeight != types.DefaultActivationHeight
	if activationInProgress && ctx.BlockHeight() >= scheme.ActivationHeight {
		err = k.JailValidators(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}

		scheme.IsActivated = true
		scheme.ActivationHeight = types.DefaultActivationHeight
		err = k.SetProvingScheme(ctx, scheme)
		if err != nil {
			return err
		}

		k.Logger(ctx).Info("proving scheme activated", "key_index", utils.SEDAKeyIndexSecp256k1)
		return
	}

	// Check the public key registration rate and start the activation
	// process if the rate has reached the threshold. If the activation
	// process is already in progress and the threshold is not met,
	// cancel the activation process.
	met, err := k.CheckKeyRegistrationRate(ctx, utils.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	if !activationInProgress && met {
		err = k.StartProvingSchemeActivation(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
	} else if activationInProgress && !met {
		err = k.CancelProvingSchemeActivation(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
	}
	return
}

// CheckKeyRegistrationRate checks if the current registration rate of
// public keys of the given key scheme exceeds the threshold.
func (k Keeper) CheckKeyRegistrationRate(ctx sdk.Context, keyIndex utils.SEDAKeyIndex) (bool, error) {
	// If the sum of the voting power has reached 80%, enable secp256k1
	// proving scheme.
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return false, err
	}

	var powerSum uint64
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		_, err := k.GetValidatorKeyAtIndex(ctx, valAddr, keyIndex)
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

	activationThresholdPercent, err := k.GetActivationThresholdPercent(ctx)
	if err != nil {
		return false, err
	}

	requiredPower := totalPower.Uint64()*uint64(activationThresholdPercent) + 1
	gotPower := powerSum * 100

	k.Logger(ctx).Info("checked status of secp256k1 proving scheme", "required", requiredPower, "got", gotPower)

	if gotPower >= requiredPower {
		return true, nil
	}
	return false, nil
}

// JailValidators goes through all validators in the store and jails
// validators without the public key corresponding to the given key
// scheme.
func (k Keeper) JailValidators(ctx sdk.Context, keyIndex utils.SEDAKeyIndex) error {
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	for _, val := range validators {
		valAddr, err := k.validatorAddressCodec.StringToBytes(val.OperatorAddress)
		if err != nil {
			return err
		}
		_, err = k.GetValidatorKeyAtIndex(ctx, valAddr, keyIndex)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				consAddr, err := val.GetConsAddr()
				if err != nil {
					return err
				}
				err = k.slashingKeeper.Jail(ctx, consAddr)
				if err != nil {
					return err
				}
				k.Logger(ctx).Info(
					"jailed validator for not having required public key",
					"consensus_address", consAddr,
					"operator_address", val.OperatorAddress,
					"key_index", keyIndex,
				)
			} else {
				return err
			}
		}
	}
	return nil
}
