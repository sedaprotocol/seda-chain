package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

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

		if timeoutHeight > uint64(ctx.BlockHeight()) {
			break
		}

		// Update data request status to tallying.
		dr, err := k.DataRequests.Get(ctx, drID)
		if err != nil {
			return err
		}

		dr.Status = types.DATA_REQUEST_TALLYING

		err = k.RevealingToTallying(ctx, dr.Index())
		if err != nil {
			return err
		}
		err = k.DataRequests.Set(ctx, drID, dr)
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
