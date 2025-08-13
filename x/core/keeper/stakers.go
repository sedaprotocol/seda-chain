package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// GetStaker retrieves a staker given its public key.
func (k Keeper) GetStaker(ctx sdk.Context, pubKey string) (types.Staker, error) {
	staker, err := k.Stakers.Get(ctx, pubKey)
	if err != nil {
		return types.Staker{}, err
	}
	return staker, nil
}

// SetStaker sets a staker in the store.
func (k Keeper) SetStaker(ctx sdk.Context, staker types.Staker) error {
	return k.Stakers.Set(ctx, staker.PublicKey, staker)
}

// GetStakersCount returns the number of stakers in the store.
func (k Keeper) GetStakersCount(ctx sdk.Context) (int, error) {
	count := 0
	err := k.Stakers.Walk(ctx, nil, func(key string, value types.Staker) (stop bool, err error) {
		count++
		return false, nil
	})
	return count, err
}
