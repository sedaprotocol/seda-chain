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
