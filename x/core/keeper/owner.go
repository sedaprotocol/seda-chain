package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) SetOwner(ctx sdk.Context) error {
	// get the new owner from the pending owner
	pendingOwner, err := k.pendingOwner.Get(ctx)
	if err != nil {
		return err
	}
	// set the new owner
	err = k.owner.Set(ctx, pendingOwner)
	if err != nil {
		return err
	}
	// clear the pending owner
	return k.pendingOwner.Remove(ctx)
}

func (k Keeper) GetPendingOwner(ctx sdk.Context) (string, error) {
	pendingOwner, err := k.pendingOwner.Get(ctx)
	if err != nil {
		return "", err
	}
	return pendingOwner, nil
}

func (k Keeper) SetPendingOwner(ctx sdk.Context, pendingOwner string) error {
	_, err := sdk.AccAddressFromBech32(pendingOwner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid pending owner address: %s", pendingOwner)
	}
	return k.pendingOwner.Set(ctx, pendingOwner)
}

func (k Keeper) IsAllowlisted(ctx sdk.Context, pubKey string) (bool, error) {
	return k.allowlist.Has(ctx, pubKey)
}

func (k Keeper) AddToAllowlist(ctx sdk.Context, pubKey string) error {
	return k.allowlist.Set(ctx, pubKey)
}

func (k Keeper) RemoveFromAllowlist(ctx sdk.Context, pubKey string) error {
	return k.allowlist.Remove(ctx, pubKey)
}

func (k Keeper) IsPaused(ctx sdk.Context) (bool, error) {
	paused, err := k.paused.Get(ctx)
	if err != nil {
		return false, err
	}
	return paused, nil
}

func (k Keeper) Pause(ctx sdk.Context) error {
	return k.paused.Set(ctx, true)
}

func (k Keeper) Unpause(ctx sdk.Context) error {
	return k.paused.Set(ctx, false)
}

func (k Keeper) GetOwner(ctx sdk.Context) (string, error) {
	owner, err := k.owner.Get(ctx)
	if err != nil {
		return "", err
	}
	return owner, nil
}
