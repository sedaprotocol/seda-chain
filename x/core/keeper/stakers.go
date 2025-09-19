package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// GetStaker retrieves a staker given its public key.
func (k Keeper) GetStaker(ctx sdk.Context, pubKey string) (types.Staker, error) {
	staker, err := k.stakers.Get(ctx, pubKey)
	if err != nil {
		return types.Staker{}, err
	}
	return staker, nil
}

// GetExecutors retrieves a list of stakers in the order of their index.
// starting at the offset and
func (k Keeper) GetExecutors(ctx sdk.Context, offset, limit uint32) ([]types.Staker, error) {
	return k.stakers.GetExecutors(ctx, offset, limit)
}

// SetStaker sets a staker in the store.
func (k Keeper) SetStaker(ctx sdk.Context, staker types.Staker) error {
	return k.stakers.Set(ctx, staker.PublicKey, staker)
}

// RemoveStaker removes a staker from the store.
func (k Keeper) RemoveStaker(ctx sdk.Context, pubKey string) error {
	return k.stakers.Remove(ctx, pubKey)
}

// GetStakerCount returns the number of stakers in the store.
func (k Keeper) GetStakerCount(ctx sdk.Context) (uint32, error) {
	return k.stakers.GetStakerCount(ctx)
}

// GetStakerIndex returns the index of a staker given its public key.
func (k Keeper) GetStakerIndex(ctx sdk.Context, pubKey string) (uint32, error) {
	return k.stakers.GetStakerIndex(ctx, pubKey)
}

// GetStakerKey returns the public key of a staker given its index.
func (k Keeper) GetStakerKey(ctx sdk.Context, index uint32) (string, error) {
	return k.stakers.GetStakerKey(ctx, index)
}

func (k Keeper) GetAllStakers(ctx sdk.Context) ([]types.Staker, error) {
	var stakers []types.Staker
	err := k.IterateStakers(ctx, func(staker types.Staker) bool {
		stakers = append(stakers, staker)
		return false
	})
	if err != nil {
		return nil, err
	}
	return stakers, nil
}

func (k Keeper) IterateStakers(ctx sdk.Context, callback func(types.Staker) (stop bool)) error {
	iter, err := k.stakers.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if callback(kv.Value) {
			break
		}
	}
	return nil
}
