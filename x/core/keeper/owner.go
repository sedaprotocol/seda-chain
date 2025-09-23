package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Get owner address; returns empty string if no owner is set
func (k Keeper) GetOwner(ctx sdk.Context) (string, error) {
	owner, err := k.owner.Get(ctx)
	if err != nil {
		return "", err
	}
	return owner, nil
}

// Set owner to the given address after validating it's a proper address
func (k Keeper) SetOwner(ctx sdk.Context, newOwner string) error {
	err := k.owner.Set(ctx, newOwner)
	if err != nil {
		return err
	}
	return nil
}

// Get pending owner address; returns empty string if no pending owner is set
func (k Keeper) GetPendingOwner(ctx sdk.Context) (string, error) {
	pendingOwner, err := k.pendingOwner.Get(ctx)
	if err != nil {
		return "", err
	}
	return pendingOwner, nil
}

// Set pending owner to the given address after validating it's a proper address
func (k Keeper) SetPendingOwner(ctx sdk.Context, pendingOwner string) error {
	return k.pendingOwner.Set(ctx, pendingOwner)
}

// IsAllowlisted checks if the given public key is in the allowlist.
func (k Keeper) IsAllowlisted(ctx sdk.Context, pubKey string) (bool, error) {
	return k.allowlist.Has(ctx, pubKey)
}

// AddToAllowlist adds the given public key to the allowlist.
func (k Keeper) AddToAllowlist(ctx sdk.Context, pubKey string) error {
	return k.allowlist.Set(ctx, pubKey)
}

// RemoveFromAllowlist removes the given public key from the allowlist.
func (k Keeper) RemoveFromAllowlist(ctx sdk.Context, pubKey string) error {
	return k.allowlist.Remove(ctx, pubKey)
}

// ListAllowlist retrieves a list of all public keys in the allowlist.
func (k Keeper) ListAllowlist(ctx sdk.Context) ([]string, error) {
	iter, err := k.allowlist.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	keys, err := iter.Keys()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Checks if the module is paused
func (k Keeper) IsPaused(ctx sdk.Context) (bool, error) {
	paused, err := k.paused.Get(ctx)
	if err != nil {
		return false, err
	}
	return paused, nil
}

// Pause the module
func (k Keeper) Pause(ctx sdk.Context) error {
	return k.paused.Set(ctx, true)
}

// Unpause the module
func (k Keeper) Unpause(ctx sdk.Context) error {
	return k.paused.Set(ctx, false)
}
