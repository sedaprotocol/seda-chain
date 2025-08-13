package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) AddToCommitting(ctx sdk.Context, index types.DataRequestIndex) error {
	return k.committing.Set(ctx, index)
}

func (k Keeper) CommittingToRevealing(ctx sdk.Context, index types.DataRequestIndex) error {
	exists, err := k.committing.Has(ctx, index)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("data request %s not found in committing", index)
	}
	err = k.committing.Remove(ctx, index)
	if err != nil {
		return err
	}
	return k.revealing.Set(ctx, index)
}

func (k Keeper) RevealingToTallying(ctx sdk.Context, index types.DataRequestIndex) error {
	exists, err := k.revealing.Has(ctx, index)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("data request %s not found in revealing", index)
	}
	err = k.revealing.Remove(ctx, index)
	if err != nil {
		return err
	}
	return k.tallying.Set(ctx, index)
}

func (k Keeper) RemoveFromTallying(ctx sdk.Context, index types.DataRequestIndex) error {
	return k.tallying.Remove(ctx, index)
}

func (k Keeper) GetTallyingDataRequestIDs(ctx sdk.Context) ([]string, error) {
	iter, err := k.tallying.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var ids []string
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		ids = append(ids, key.DrID())
	}
	return ids, nil
}
