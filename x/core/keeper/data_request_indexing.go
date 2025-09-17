package keeper

import (
	"cosmossdk.io/collections"
	collcdc "cosmossdk.io/collections/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type DataRequestIndexSet struct {
	collections.KeySet[collections.Pair[types.DataRequestStatus, types.DataRequestIndex]]
	committingCount collections.Item[uint64]
	revealingCount  collections.Item[uint64]
	tallyingCount   collections.Item[uint64]
}

func NewDataRequestIndexSet(sb *collections.SchemaBuilder) DataRequestIndexSet {
	return DataRequestIndexSet{
		KeySet:          collections.NewKeySet(sb, types.DataRequestIndexingPrefix, "data_request_indexing", collections.PairKeyCodec(collcdc.NewInt32Key[types.DataRequestStatus](), collcdc.NewBytesKey[types.DataRequestIndex]())),
		committingCount: collections.NewItem(sb, types.DataRequestCommittingKey, "committing_count", collections.Uint64Value),
		revealingCount:  collections.NewItem(sb, types.DataRequestRevealingKey, "revealing_count", collections.Uint64Value),
		tallyingCount:   collections.NewItem(sb, types.DataRequestTallyingKey, "tallying_count", collections.Uint64Value),
	}
}

// Set overrides the KeySet method to track data request count by status as well.
func (s DataRequestIndexSet) Set(ctx sdk.Context, key collections.Pair[types.DataRequestStatus, types.DataRequestIndex]) error {
	count, err := s.getDataRequestCountByStatus(ctx, key.K1())
	if err != nil {
		return err
	}
	err = s.KeySet.Set(ctx, key)
	if err != nil {
		return err
	}
	return s.setDataRequestCountByStatus(ctx, key.K1(), count+1)
}

// Remove overrides the KeySet Remove to track data request count by status as well.
func (s DataRequestIndexSet) Remove(ctx sdk.Context, key collections.Pair[types.DataRequestStatus, types.DataRequestIndex]) error {
	count, err := s.getDataRequestCountByStatus(ctx, key.K1())
	if err != nil {
		return err
	}
	err = s.KeySet.Remove(ctx, key)
	if err != nil {
		return err
	}
	if count == 0 {
		return types.ErrUnexpectDataRequestCount.Wrapf("count is 0, status %s", key.K1())
	}
	return s.setDataRequestCountByStatus(ctx, key.K1(), count-1)
}

func (s DataRequestIndexSet) getDataRequestCountByStatus(ctx sdk.Context, status types.DataRequestStatus) (uint64, error) {
	switch status {
	case types.DATA_REQUEST_STATUS_COMMITTING:
		return s.committingCount.Get(ctx)
	case types.DATA_REQUEST_STATUS_REVEALING:
		return s.revealingCount.Get(ctx)
	case types.DATA_REQUEST_STATUS_TALLYING:
		return s.tallyingCount.Get(ctx)
	default:
		return 0, types.ErrInvalidDataRequestStatus.Wrapf("invalid status: %s", status)
	}
}

func (s DataRequestIndexSet) setDataRequestCountByStatus(ctx sdk.Context, status types.DataRequestStatus, count uint64) error {
	switch status {
	case types.DATA_REQUEST_STATUS_COMMITTING:
		return s.committingCount.Set(ctx, count)
	case types.DATA_REQUEST_STATUS_REVEALING:
		return s.revealingCount.Set(ctx, count)
	case types.DATA_REQUEST_STATUS_TALLYING:
		return s.tallyingCount.Set(ctx, count)
	default:
		return types.ErrInvalidDataRequestStatus.Wrapf("invalid status: %s", status)
	}
}

// GetDataRequestsByStatus returns data requests by status. You can specify the
// limit of the query by proving a non-zero limit. If you provide a non-nil
// lastSeenIndex, the query will start from the next data request after the
// lastSeenIndex. The method also returns the new lastSeenIndex and the total
// number of data requests under the given status.
func (k Keeper) GetDataRequestsByStatus(ctx sdk.Context, status types.DataRequestStatus, limit uint64, lastSeenIndex types.DataRequestIndex) ([]types.DataRequest, types.DataRequestIndex, uint64, error) {
	total, err := k.dataRequestIndexing.getDataRequestCountByStatus(ctx, status)
	if err != nil {
		return nil, nil, 0, err
	}

	rng := collections.NewPrefixedPairRange[types.DataRequestStatus, types.DataRequestIndex](status)
	if lastSeenIndex != nil {
		rng.StartExclusive(lastSeenIndex)
	}

	iter, err := k.dataRequestIndexing.Iterate(ctx, rng)
	if err != nil {
		return nil, nil, 0, err
	}
	defer iter.Close()

	var dataRequests []types.DataRequest
	var newLastSeenIndex types.DataRequestIndex
	count := uint64(0)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, nil, 0, err
		}

		dataRequest, err := k.GetDataRequest(ctx, key.K2().DrID())
		if err != nil {
			return nil, nil, 0, err
		}

		dataRequests = append(dataRequests, dataRequest)
		newLastSeenIndex = key.K2()
		count++
		if count >= limit {
			break
		}
	}
	return dataRequests, newLastSeenIndex, total, nil
}

func (k Keeper) GetDataRequestIDsByStatus(ctx sdk.Context, status types.DataRequestStatus) ([]string, error) {
	rng := collections.NewPrefixedPairRange[types.DataRequestStatus, types.DataRequestIndex](status)

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
