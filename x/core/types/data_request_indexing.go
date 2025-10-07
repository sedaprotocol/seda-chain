package types

import (
	"cosmossdk.io/collections"
	collcdc "cosmossdk.io/collections/codec"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DataRequestIndexing is used to store data request objects while maintaining
// their indices to facilitate querying and counting data requests by status.
type DataRequestIndexing struct {
	dataRequests    collections.Map[string, DataRequest]
	indexing        collections.KeySet[collections.Pair[DataRequestStatus, DataRequestIndex]]
	committingCount collections.Item[uint64]
	revealingCount  collections.Item[uint64]
	tallyingCount   collections.Item[uint64]
}

func NewDataRequestIndexing(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) DataRequestIndexing {
	return DataRequestIndexing{
		dataRequests:    collections.NewMap(sb, DataRequestsKeyPrefix, "data_requests", collections.StringKey, codec.CollValue[DataRequest](cdc)),
		indexing:        collections.NewKeySet(sb, DataRequestIndexingPrefix, "data_request_indexing", collections.PairKeyCodec(collcdc.NewInt32Key[DataRequestStatus](), collcdc.NewBytesKey[DataRequestIndex]())),
		committingCount: collections.NewItem(sb, DataRequestCommittingKey, "committing_count", collections.Uint64Value),
		revealingCount:  collections.NewItem(sb, DataRequestRevealingKey, "revealing_count", collections.Uint64Value),
		tallyingCount:   collections.NewItem(sb, DataRequestTallyingKey, "tallying_count", collections.Uint64Value),
	}
}

// GetDataRequest retrieves a data request given its hex-encoded ID.
func (s DataRequestIndexing) GetDataRequest(ctx sdk.Context, id string) (DataRequest, error) {
	return s.dataRequests.Get(ctx, id)
}

// StoreDataRequest stores a given data request in the store, while also updating
// the counter and the indexing.
func (s DataRequestIndexing) StoreDataRequest(ctx sdk.Context, dr DataRequest) error {
	// Check if a data request with the same ID already exists.
	exists, err := s.dataRequests.Has(ctx, dr.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrDataRequestAlreadyExists.Wrapf("data request ID %s", dr.ID)
	}

	// Store the data request, increment the counter, and update the indexing.
	err = s.dataRequests.Set(ctx, dr.ID, dr)
	if err != nil {
		return err
	}

	err = s.IncrementDataRequestCountByStatus(ctx, dr.Status)
	if err != nil {
		return err
	}

	return s.indexing.Set(ctx, collections.Join(dr.Status, dr.Index()))
}

// RemoveDataRequest removes a data request from the store, while also updating
// the counter and the indexing.
func (s DataRequestIndexing) RemoveDataRequest(ctx sdk.Context, index DataRequestIndex, status DataRequestStatus) error {
	indexKey := collections.Join(status, index)
	dataRequestID := index.DrID()

	exists, err := s.indexing.Has(ctx, indexKey)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDataRequestIndexNotFound.Wrapf("data request ID %s, status %s", dataRequestID, status)
	}

	err = s.DecrementDataRequestCountByStatus(ctx, status)
	if err != nil {
		return err
	}

	err = s.dataRequests.Remove(ctx, dataRequestID)
	if err != nil {
		return err
	}

	return s.indexing.Remove(ctx, indexKey)
}

// UpdateDataRequest updates an existing data request. If newStatus is not nil,
// the data request status will also be updated.
func (s DataRequestIndexing) UpdateDataRequest(ctx sdk.Context, dr *DataRequest, newStatus *DataRequestStatus) error {
	exists, err := s.dataRequests.Has(ctx, dr.ID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDataRequestNotFound.Wrapf("data request ID %s", dr.ID)
	}

	// Update counter and indexing only in case of status transition.
	if newStatus != nil {
		err := verifyStatusTransition(dr.Status, *newStatus)
		if err != nil {
			return err
		}

		// Decrement the counter and remove the current index.
		err = s.DecrementDataRequestCountByStatus(ctx, dr.Status)
		if err != nil {
			return err
		}

		exists, err = s.indexing.Has(ctx, collections.Join(dr.Status, dr.Index()))
		if err != nil {
			return err
		}
		if !exists {
			return ErrDataRequestIndexNotFound.Wrapf("data request ID %s, status %s", dr.ID, dr.Status)
		}
		err = s.indexing.Remove(ctx, collections.Join(dr.Status, dr.Index()))
		if err != nil {
			return err
		}

		// Increment the counter and set the new index.
		err = s.IncrementDataRequestCountByStatus(ctx, *newStatus)
		if err != nil {
			return err
		}

		err = s.indexing.Set(ctx, collections.Join(*newStatus, dr.Index()))
		if err != nil {
			return err
		}

		dr.Status = *newStatus
	}

	return s.dataRequests.Set(ctx, dr.ID, *dr)
}

// RemoveDataRequestIndex removes an index from the indexing, returning an error
// if the index is not found.
func (s DataRequestIndexing) RemoveDataRequestIndex(ctx sdk.Context, index DataRequestIndex, status DataRequestStatus) error {
	indexKey := collections.Join(status, index)
	exists, err := s.indexing.Has(ctx, indexKey)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDataRequestIndexNotFound.Wrapf("data request ID %s, status %s", index.DrID(), status)
	}
	return s.indexing.Remove(ctx, indexKey)
}

// GetDataRequestIDsByStatus returns the IDs of the data requests under the given status.
func (s DataRequestIndexing) GetDataRequestIDsByStatus(ctx sdk.Context, status DataRequestStatus) ([]string, error) {
	rng := collections.NewPrefixedPairRange[DataRequestStatus, DataRequestIndex](status)

	iter, err := s.indexing.Iterate(ctx, rng)
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

// GetDataRequestsByStatus returns data requests by status. You can specify the
// limit of the query by proving a non-zero limit. If you provide a non-nil
// lastSeenIndex, the query will start from the next data request after the
// lastSeenIndex. The method also returns the new lastSeenIndex and the total
// number of data requests under the given status.
func (s DataRequestIndexing) GetDataRequestsByStatus(ctx sdk.Context, status DataRequestStatus, limit uint64, lastSeenIndex DataRequestIndex) ([]DataRequest, DataRequestIndex, uint64, error) {
	total, err := s.GetDataRequestCountByStatus(ctx, status)
	if err != nil {
		return nil, nil, 0, err
	}

	rng := collections.NewPrefixedPairRange[DataRequestStatus, DataRequestIndex](status).Descending()
	if lastSeenIndex != nil {
		rng.EndExclusive(lastSeenIndex)
	}

	iter, err := s.indexing.Iterate(ctx, rng)
	if err != nil {
		return nil, nil, 0, err
	}
	defer iter.Close()

	var dataRequests []DataRequest
	var newLastSeenIndex DataRequestIndex
	count := uint64(0)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, nil, 0, err
		}

		dataRequest, err := s.dataRequests.Get(ctx, key.K2().DrID())
		if err != nil {
			return nil, nil, 0, err
		}

		dataRequests = append(dataRequests, dataRequest)
		newLastSeenIndex = append([]byte(nil), key.K2()...)
		count++
		if count >= limit {
			break
		}
	}
	return dataRequests, newLastSeenIndex, total, nil
}

// GetAllDataRequests retrieves all data requests from the store.
func (s DataRequestIndexing) GetAllDataRequests(ctx sdk.Context) ([]DataRequest, error) {
	iter, err := s.dataRequests.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	dataRequests := make([]DataRequest, 0)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		dataRequest, err := s.dataRequests.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		dataRequests = append(dataRequests, dataRequest)
	}
	return dataRequests, nil
}

// InitializeCounters initializes the counters to 0.
// Intended to be used in InitGenesis.
func (s DataRequestIndexing) InitializeCounters(ctx sdk.Context) error {
	err := s.committingCount.Set(ctx, 0)
	if err != nil {
		return err
	}
	err = s.revealingCount.Set(ctx, 0)
	if err != nil {
		return err
	}
	err = s.tallyingCount.Set(ctx, 0)
	if err != nil {
		return err
	}
	return nil
}

func (s DataRequestIndexing) IncrementDataRequestCountByStatus(ctx sdk.Context, status DataRequestStatus) error {
	count, err := s.GetDataRequestCountByStatus(ctx, status)
	if err != nil {
		return err
	}
	err = s.SetDataRequestCountByStatus(ctx, status, count+1)
	if err != nil {
		return err
	}
	return nil
}

func (s DataRequestIndexing) DecrementDataRequestCountByStatus(ctx sdk.Context, status DataRequestStatus) error {
	count, err := s.GetDataRequestCountByStatus(ctx, status)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrUnexpectDataRequestCount.Wrapf("count is 0 for status %s", status)
	}
	err = s.SetDataRequestCountByStatus(ctx, status, count-1)
	if err != nil {
		return err
	}
	return nil
}

func (s DataRequestIndexing) GetDataRequestCountByStatus(ctx sdk.Context, status DataRequestStatus) (uint64, error) {
	switch status {
	case DATA_REQUEST_STATUS_COMMITTING:
		return s.committingCount.Get(ctx)
	case DATA_REQUEST_STATUS_REVEALING:
		return s.revealingCount.Get(ctx)
	case DATA_REQUEST_STATUS_TALLYING:
		return s.tallyingCount.Get(ctx)
	default:
		return 0, ErrInvalidDataRequestStatus.Wrapf("invalid status: %s", status)
	}
}

func (s DataRequestIndexing) SetDataRequestCountByStatus(ctx sdk.Context, status DataRequestStatus, count uint64) error {
	switch status {
	case DATA_REQUEST_STATUS_COMMITTING:
		return s.committingCount.Set(ctx, count)
	case DATA_REQUEST_STATUS_REVEALING:
		return s.revealingCount.Set(ctx, count)
	case DATA_REQUEST_STATUS_TALLYING:
		return s.tallyingCount.Set(ctx, count)
	default:
		return ErrInvalidDataRequestStatus.Wrapf("invalid status: %s", status)
	}
}

// verifyStatusTransition returns an error if and only if the status transition
// is invalid. A valid status transition is: Committing -> Revealing -> Tallying,
// except that in case of a timeout, in which Committing -> Tallying is also possible.
func verifyStatusTransition(currentStatus, newStatus DataRequestStatus) error {
	switch newStatus {
	case DATA_REQUEST_STATUS_COMMITTING:
		// There can never be a transition to committing status.
		return ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
	case DATA_REQUEST_STATUS_REVEALING:
		if currentStatus != DATA_REQUEST_STATUS_COMMITTING {
			return ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
		}
	case DATA_REQUEST_STATUS_TALLYING:
		if currentStatus != DATA_REQUEST_STATUS_COMMITTING &&
			currentStatus != DATA_REQUEST_STATUS_REVEALING {
			return ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
		}
	default: //	case DATA_REQUEST_STATUS_UNSPECIFIED:
		return ErrInvalidStatusTransition.Wrapf("%s -> %s", currentStatus, newStatus)
	}
	return nil
}
