package keeper

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
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

	scheme, err := k.GetProvingScheme(ctx, sedatypes.SEDAKeyIndexSecp256k1)
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
		err = k.JailValidators(ctx, sedatypes.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}

		scheme.IsActivated = true
		scheme.ActivationHeight = types.DefaultActivationHeight
		err = k.SetProvingScheme(ctx, scheme)
		if err != nil {
			return err
		}

		k.Logger(ctx).Info("proving scheme activated", "key_index", sedatypes.SEDAKeyIndexSecp256k1)
		return
	}

	// Check the public key registration rate and start the activation
	// process if the rate has reached the threshold. If the activation
	// process is already in progress and the threshold is not met,
	// cancel the activation process.
	met, err := k.CheckKeyRegistrationRate(ctx, sedatypes.SEDAKeyIndexSecp256k1)
	if err != nil {
		return err
	}
	if !activationInProgress && met {
		err = k.StartProvingSchemeActivation(ctx, sedatypes.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
	} else if activationInProgress && !met {
		err = k.CancelProvingSchemeActivation(ctx, sedatypes.SEDAKeyIndexSecp256k1)
		if err != nil {
			return err
		}
	}
	return
}

// CheckKeyRegistrationRate checks if the current registration rate of
// public keys of the given key scheme exceeds the threshold.
func (k Keeper) CheckKeyRegistrationRate(ctx sdk.Context, keyIndex sedatypes.SEDAKeyIndex) (bool, error) {
	// If the sum of the voting power has reached 80%, enable secp256k1
	// proving scheme.
	totalPower, err := k.stakingKeeper.GetLastTotalPower(ctx)
	if err != nil {
		return false, err
	}

	powerSum := math.ZeroInt()
	err = k.stakingKeeper.IterateLastValidatorPowers(ctx, func(valAddr sdk.ValAddress, power int64) (stop bool) {
		registered, err := k.HasRegisteredKey(ctx, valAddr, keyIndex)
		if err != nil {
			panic(err)
		}
		if !registered {
			return false
		}
		//nolint:gosec // G115: We shouldn't get negative power anyway.
		powerSum = powerSum.Add(math.NewInt(power))
		return false
	})
	if err != nil {
		return false, err
	}

	activationThresholdPercent, err := k.GetActivationThresholdPercent(ctx)
	if err != nil {
		return false, err
	}

	requiredPower := totalPower.Mul(math.NewIntFromUint64(uint64(activationThresholdPercent))).Add(math.OneInt())
	gotPower := powerSum.Mul(math.NewInt(100))

	k.Logger(ctx).Info("checked status of secp256k1 proving scheme",
		"required", requiredPower.String(), "got", gotPower.String())

	if gotPower.GTE(requiredPower) {
		return true, nil
	}
	return false, nil
}

// JailValidators goes through all validators in the store and jails
// validators without the public key corresponding to the given key
// scheme.
func (k Keeper) JailValidators(ctx sdk.Context, keyIndex sedatypes.SEDAKeyIndex) error {
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	for _, val := range validators {
		valAddr, err := k.validatorAddressCodec.StringToBytes(val.OperatorAddress)
		if err != nil {
			return err
		}
		registered, err := k.HasRegisteredKey(ctx, valAddr, keyIndex)
		if err != nil {
			return err
		}
		if !registered {
			consAddr, err := val.GetConsAddr()
			if err != nil {
				return err
			}
			if !val.IsJailed() {
				err = k.slashingKeeper.Jail(ctx, consAddr)
				if err != nil {
					return err
				}
			}
			k.Logger(ctx).Info(
				"validator is jailed for missing required public key (or was already jailed)",
				"consensus_address", consAddr,
				"operator_address", val.OperatorAddress,
				"key_index", keyIndex,
			)
		}
	}

	return nil
}
