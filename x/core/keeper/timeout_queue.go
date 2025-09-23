package keeper

import (
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) AddToTimeoutQueue(ctx sdk.Context, drID string, timeoutHeight int64) error {
	return k.timeoutQueue.Set(ctx, collections.Join(timeoutHeight, drID))
}

func (k Keeper) RemoveFromTimeoutQueue(ctx sdk.Context, drID string, timeoutHeight int64) error {
	err := k.timeoutQueue.Remove(ctx, collections.Join(timeoutHeight, drID))
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateDataRequestTimeout(ctx sdk.Context, drID string, oldTimeoutHeight, newTimeoutHeight int64) error {
	exists, err := k.timeoutQueue.Has(ctx, collections.Join(oldTimeoutHeight, drID))
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("data request %s not found in timeout queue", drID)
	}
	err = k.timeoutQueue.Remove(ctx, collections.Join(oldTimeoutHeight, drID))
	if err != nil {
		return err
	}
	return k.timeoutQueue.Set(ctx, collections.Join(newTimeoutHeight, drID))
}

func (k Keeper) ExpireDataRequests(ctx sdk.Context) error {
	iter, err := k.timeoutQueue.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		timeoutHeight := key.K1()
		drID := key.K2()

		if timeoutHeight > ctx.BlockHeight() {
			continue
		}

		// Update data request status to tallying.
		dr, err := k.GetDataRequest(ctx, drID)
		if err != nil {
			return err
		}

		err = k.UpdateDataRequestIndexing(ctx, &dr, types.DATA_REQUEST_STATUS_TALLYING)
		if err != nil {
			return err
		}
		dr.TimeoutHeight = -1

		err = k.SetDataRequest(ctx, dr)
		if err != nil {
			return err
		}

		// Remove from timeout queue.
		err = k.timeoutQueue.Remove(ctx, key)
		if err != nil {
			return err
		}
		k.Logger(ctx).Debug("expired data request", "ID", drID)
	}

	return nil
}

func (k Keeper) ListTimeoutQueue(ctx sdk.Context) ([]collections.Pair[int64, string], error) {
	iter, err := k.timeoutQueue.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	return iter.Keys()
}
