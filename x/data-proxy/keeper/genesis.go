package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

// InitGenesis initializes the store based on the given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, dataProxyConfig := range data.DataProxyConfigs {
		if err := k.SetDataProxyConfig(ctx, dataProxyConfig.DataProxyPubkey, *dataProxyConfig.Config); err != nil {
			panic(err)
		}
	}

	for _, feeUpdate := range data.FeeUpdateQueue {
		if err := k.SetFeeUpdate(ctx, feeUpdate.UpdateHeight, feeUpdate.DataProxyPubkey); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis extracts all data from store to genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	var gs types.GenesisState

	params, err := k.params.Get(ctx)
	if err != nil {
		panic(err)
	}
	gs.Params = params

	configs, err := k.getAllDataProxyConfigs(ctx)
	if err != nil {
		panic(err)
	}
	gs.DataProxyConfigs = configs

	feeUpdates, err := k.getAllFeeUpdateRecords(ctx)
	if err != nil {
		panic(err)
	}
	gs.FeeUpdateQueue = feeUpdates

	return gs
}

func (k Keeper) getAllDataProxyConfigs(ctx sdk.Context) ([]types.DataProxyConfig, error) {
	configs := make([]types.DataProxyConfig, 0)

	itr, err := k.dataProxyConfigs.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		kv, err := itr.KeyValue()
		if err != nil {
			return nil, err
		}

		configs = append(configs, types.DataProxyConfig{
			DataProxyPubkey: kv.Key,
			Config:          &kv.Value,
		})
	}

	return configs, nil
}

func (k Keeper) getAllFeeUpdateRecords(ctx sdk.Context) ([]types.FeeUpdateQueueRecord, error) {
	feeUpdates := make([]types.FeeUpdateQueueRecord, 0)

	itr, err := k.feeUpdateQueue.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key, err := itr.Key()
		if err != nil {
			return nil, err
		}

		feeUpdates = append(feeUpdates, types.FeeUpdateQueueRecord{
			UpdateHeight:    key.K1(),
			DataProxyPubkey: key.K2(),
		})
	}

	return feeUpdates, nil
}
