package keeper

import (
	"cosmossdk.io/collections"
	collcdc "cosmossdk.io/collections/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// DataRequestIndexing is used to maintain data request indexing by status and
// track counts of data requests by status.
type DataRequestIndexing struct {
	collections.KeySet[collections.Pair[types.DataRequestStatus, types.DataRequestIndex]]
	committingCount collections.Item[uint64]
	revealingCount  collections.Item[uint64]
	tallyingCount   collections.Item[uint64]
}

func NewDataRequestIndexing(sb *collections.SchemaBuilder) DataRequestIndexing {
	return DataRequestIndexing{
		KeySet:          collections.NewKeySet(sb, types.DataRequestIndexingPrefix, "data_request_indexing", collections.PairKeyCodec(collcdc.NewInt32Key[types.DataRequestStatus](), collcdc.NewBytesKey[types.DataRequestIndex]())),
		committingCount: collections.NewItem(sb, types.DataRequestCommittingKey, "committing_count", collections.Uint64Value),
		revealingCount:  collections.NewItem(sb, types.DataRequestRevealingKey, "revealing_count", collections.Uint64Value),
		tallyingCount:   collections.NewItem(sb, types.DataRequestTallyingKey, "tallying_count", collections.Uint64Value),
	}
}

// Set overrides the KeySet method to track data request count by status.
func (s DataRequestIndexing) Set(ctx sdk.Context, key collections.Pair[types.DataRequestStatus, types.DataRequestIndex]) error {
	exists, err := s.Has(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		count, err := s.getDataRequestCountByStatus(ctx, key.K1())
		if err != nil {
			return err
		}
		err = s.setDataRequestCountByStatus(ctx, key.K1(), count+1)
		if err != nil {
			return err
		}
	}
	return s.KeySet.Set(ctx, key)
}

// Remove overrides the KeySet Remove to track data request count by status.
func (s DataRequestIndexing) Remove(ctx sdk.Context, key collections.Pair[types.DataRequestStatus, types.DataRequestIndex]) error {
	exists, err := s.Has(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		return types.ErrDataRequestStatusNotFound.Wrapf("data request ID %s, status %s", key.K2().DrID(), key.K1())
	}

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

func (s DataRequestIndexing) getDataRequestCountByStatus(ctx sdk.Context, status types.DataRequestStatus) (uint64, error) {
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

func (s DataRequestIndexing) setDataRequestCountByStatus(ctx sdk.Context, status types.DataRequestStatus, count uint64) error {
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

// GetDataRequestIDsByStatus returns the IDs of the data requests under the given status.
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

func (k Keeper) UpdateDataRequestIndexing(ctx sdk.Context, dr *types.DataRequest, newStatus types.DataRequestStatus) error {
	// Check the logic of the status transition while removing the index from
	// the current status set.
	// Committing -> Revealing -> Tallying, except that in case of timeout,
	// Committing -> Tallying is also possible.
	switch newStatus {
	case types.DATA_REQUEST_STATUS_COMMITTING:
		// The data request must be new.
		if dr.Status != types.DATA_REQUEST_STATUS_UNSPECIFIED {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", dr.Status, newStatus)
		}

	case types.DATA_REQUEST_STATUS_REVEALING:
		// The data request must be in committing status.
		if dr.Status != types.DATA_REQUEST_STATUS_COMMITTING {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", dr.Status, newStatus)
		}
		err := k.RemoveDataRequestIndexing(ctx, dr.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		if err != nil {
			return err
		}

	case types.DATA_REQUEST_STATUS_TALLYING:
		// The data request must be in committing or revealing status.
		if dr.Status != types.DATA_REQUEST_STATUS_COMMITTING &&
			dr.Status != types.DATA_REQUEST_STATUS_REVEALING {
			return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", dr.Status, newStatus)
		}
		err := k.RemoveDataRequestIndexing(ctx, dr.Index(), dr.Status)
		if err != nil {
			return err
		}

	default: //	case types.DATA_REQUEST_STATUS_UNSPECIFIED:
		return types.ErrInvalidStatusTransition.Wrapf("%s -> %s", dr.Status, newStatus)
	}

	err := k.dataRequestIndexing.Set(ctx, collections.Join(newStatus, dr.Index()))
	if err != nil {
		return err
	}
	dr.Status = newStatus

	return nil
}

func (k Keeper) CheckDataRequestIndexing(ctx sdk.Context, index types.DataRequestIndex, status types.DataRequestStatus) (bool, error) {
	return k.dataRequestIndexing.Has(ctx, collections.Join(status, index))
}

func (k Keeper) RemoveDataRequestIndexing(ctx sdk.Context, index types.DataRequestIndex, status types.DataRequestStatus) error {
	return k.dataRequestIndexing.Remove(ctx, collections.Join(status, index))
}
