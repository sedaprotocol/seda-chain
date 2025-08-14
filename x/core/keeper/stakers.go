package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) isAllowlisted(ctx sdk.Context, pubKey string) (bool, error) {
	return k.allowlist.Has(ctx, pubKey)
}

func (k Keeper) addToAllowlist(ctx sdk.Context, pubKey string) error {
	return k.allowlist.Set(ctx, pubKey)
}

// GetStaker retrieves a staker given its public key.
func (k Keeper) GetStaker(ctx sdk.Context, pubKey string) (types.Staker, error) {
	staker, err := k.stakers.Get(ctx, pubKey)
	if err != nil {
		return types.Staker{}, err
	}
	return staker, nil
}

// SetStaker sets a staker in the store.
func (k Keeper) SetStaker(ctx sdk.Context, staker types.Staker) error {
	return k.stakers.Set(ctx, staker.PublicKey, staker)
}

// GetStakersCount returns the number of stakers in the store.
func (k Keeper) GetStakersCount(ctx sdk.Context) (uint32, error) {
	count := uint32(0)
	err := k.stakers.Walk(ctx, nil, func(_ string, _ types.Staker) (stop bool, err error) {
		count++
		return false, nil
	})
	return count, err
}
