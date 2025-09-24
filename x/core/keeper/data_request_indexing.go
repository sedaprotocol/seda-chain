package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) StoreDataRequest(ctx sdk.Context, dr types.DataRequest) error {
	return k.dataRequests.StoreDataRequest(ctx, dr)
}

func (k Keeper) GetDataRequest(ctx sdk.Context, id string) (types.DataRequest, error) {
	return k.dataRequests.GetDataRequest(ctx, id)
}

func (k Keeper) UpdateDataRequest(ctx sdk.Context, dr *types.DataRequest, newStatus *types.DataRequestStatus) error {
	return k.dataRequests.UpdateDataRequest(ctx, dr, newStatus)
}

func (k Keeper) RemoveDataRequest(ctx sdk.Context, index types.DataRequestIndex, status types.DataRequestStatus) error {
	return k.dataRequests.RemoveDataRequest(ctx, index, status)
}

func (k Keeper) GetDataRequestsByStatus(ctx sdk.Context, status types.DataRequestStatus, limit uint64, lastSeenIndex types.DataRequestIndex) ([]types.DataRequest, types.DataRequestIndex, uint64, error) {
	return k.dataRequests.GetDataRequestsByStatus(ctx, status, limit, lastSeenIndex)
}

func (k Keeper) GetDataRequestIDsByStatus(ctx sdk.Context, status types.DataRequestStatus) ([]string, error) {
	return k.dataRequests.GetDataRequestIDsByStatus(ctx, status)
}

func (k Keeper) GetDataRequestStatus(ctx sdk.Context, id string) (types.DataRequestStatus, error) {
	dr, err := k.GetDataRequest(ctx, id)
	if err != nil {
		return types.DATA_REQUEST_STATUS_UNSPECIFIED, err
	}
	return dr.Status, nil
}

// GetDataRequestStatuses returns the statuses of the data requests given their IDs.
func (k Keeper) GetDataRequestStatuses(ctx sdk.Context, ids []string) (map[string]types.DataRequestStatus, error) {
	statuses := make(map[string]types.DataRequestStatus)
	for _, id := range ids {
		status, err := k.GetDataRequestStatus(ctx, id)
		if err != nil {
			return nil, err
		}
		statuses[id] = status
	}
	return statuses, nil
}

func (k Keeper) GetAllDataRequests(ctx sdk.Context) ([]types.DataRequest, error) {
	return k.dataRequests.GetAllDataRequests(ctx)
}
