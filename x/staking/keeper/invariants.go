package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/sedaprotocol/seda-chain/app/utils"
)

// RegisterInvariants registers all staking invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-accounts",
		keeper.ModuleAccountInvariants(k.Keeper))
	ir.RegisterRoute(types.ModuleName, "nonnegative-power",
		keeper.NonNegativePowerInvariant(k.Keeper))
	ir.RegisterRoute(types.ModuleName, "positive-delegation",
		keeper.PositiveDelegationInvariant(k.Keeper))
	ir.RegisterRoute(types.ModuleName, "delegator-shares",
		keeper.DelegatorSharesInvariant(k.Keeper))
	// Custom invariant
	ir.RegisterRoute(types.ModuleName, "seda-pubkey-registration",
		PubKeyRegistrationInvariant(k))
}

// PubKeyRegistrationInvariant checks for the invariant that once
// the secp256k1 proving scheme is enabled, all validators have
// registered their public keys.
func PubKeyRegistrationInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken bool
		var violator string
		activated, err := k.pubKeyKeeper.IsProvingSchemeActivated(ctx, utils.SEDAKeyIndexSecp256k1)
		if err != nil {
			panic(err)
		}
		if activated {
			err = k.IterateValidators(ctx, func(_ int64, validator types.ValidatorI) bool {
				valAddr, err := k.validatorAddressCodec.StringToBytes(validator.GetOperator())
				if err != nil {
					panic(err)
				}
				_, err = k.pubKeyKeeper.GetValidatorKeyAtIndex(ctx, valAddr, utils.SEDAKeyIndexSecp256k1)
				if err != nil {
					if errors.Is(err, collections.ErrNotFound) {
						broken = true
						violator = validator.GetOperator()
						return true
					}
					panic(err)
				}
				return false
			})
			if err != nil {
				panic(err)
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "SEDA public key registration",
			fmt.Sprintf("\tViolator, if any (may not be the only violator): %s\n", violator),
		), broken
	}
}
