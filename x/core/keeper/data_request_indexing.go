package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) GetTallyingDataRequestIDs(ctx sdk.Context) ([]string, error) {
	rng := collections.NewPrefixedPairRange[types.DataRequestStatus, types.DataRequestIndex](types.DATA_REQUEST_STATUS_TALLYING)

	iter, err := k.dataRequestIndexing.Iterate(ctx, rng)
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
		ids = append(ids, key.K2().DrID())
	}
	return ids, nil
}

func (k Keeper) UpdateDataRequestIndexing(ctx sdk.Context, index types.DataRequestIndex, currentStatus, newStatus types.DataRequestStatus) error {
	// Check the logic of the status transition, which follows:
	// Unspecified (addition) -> Committing -> Revealing -> Tallying -> Unspecified (removal)
	// except that in case of timeout, Committing -> Tallying is also possible.
	switch currentStatus {
	case types.DATA_REQUEST_STATUS_UNSPECIFIED:
		if newStatus != types.DATA_REQUEST_STATUS_COMMITTING {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
		}

	case types.DATA_REQUEST_STATUS_COMMITTING:
		if newStatus != types.DATA_REQUEST_STATUS_REVEALING &&
			newStatus != types.DATA_REQUEST_STATUS_TALLYING {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
		}

	case types.DATA_REQUEST_STATUS_REVEALING:
		if newStatus != types.DATA_REQUEST_STATUS_TALLYING {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
		}

	case types.DATA_REQUEST_STATUS_TALLYING:
		if newStatus != types.DATA_REQUEST_STATUS_UNSPECIFIED {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
		}
	}

	if currentStatus != types.DATA_REQUEST_STATUS_UNSPECIFIED {
		exists, err := k.dataRequestIndexing.Has(ctx, collections.Join(currentStatus, index))
		if err != nil {
			return err
		}
		if !exists {
			return types.ErrDataRequestNotFoundInIndex.Wrapf("data request ID %s, status %s", index.DrID(), currentStatus)
		}
		err = k.dataRequestIndexing.Remove(ctx, collections.Join(currentStatus, index))
		if err != nil {
			return err
		}
	}

	if newStatus != types.DATA_REQUEST_STATUS_UNSPECIFIED {
		err := k.dataRequestIndexing.Set(ctx, collections.Join(newStatus, index))
		if err != nil {
			return err
		}
	}
	return nil
}
