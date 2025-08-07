package keeper

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type Keeper struct {
	wasmStorageKeeper types.WasmStorageKeeper
	batchingKeeper    types.BatchingKeeper
	stakingKeeper     types.StakingKeeper
	bankKeeper        types.BankKeeper
	wasmKeeper        wasmtypes.ContractOpsKeeper
	wasmViewKeeper    wasmtypes.ViewKeeper
	authority         string

	Schema    collections.Schema
	Allowlist collections.KeySet[string]
	Stakers   collections.Map[string, types.Staker]
	Params    collections.Item[types.Params]

	DataRequests collections.Map[string, types.DataRequest]
	committing   collections.KeySet[types.DataRequestIndex]
	revealing    collections.KeySet[types.DataRequestIndex]
	tallying     collections.KeySet[types.DataRequestIndex]
	timeoutQueue collections.KeySet[collections.Pair[uint64, string]]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, wsk types.WasmStorageKeeper, batk types.BatchingKeeper, sk types.StakingKeeper, bank types.BankKeeper, wk wasmtypes.ContractOpsKeeper, wvk wasmtypes.ViewKeeper, authority string) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		wasmStorageKeeper: wsk,
		batchingKeeper:    batk,
		stakingKeeper:     sk,
		bankKeeper:        bank,
		wasmKeeper:        wk,
		wasmViewKeeper:    wvk,
		authority:         authority,
		Allowlist:         collections.NewKeySet(sb, types.AllowlistKey, "allowlist", collections.StringKey),
		Stakers:           collections.NewMap(sb, types.StakersKeyPrefix, "stakers", collections.StringKey, codec.CollValue[types.Staker](cdc)),
		Params:            collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		DataRequests:      collections.NewMap(sb, types.DataRequestsKeyPrefix, "data_requests", collections.StringKey, codec.CollValue[types.DataRequest](cdc)),
		committing:        collections.NewKeySet(sb, types.CommittingKeyPrefix, "committing", collcodec.NewBytesKey[types.DataRequestIndex]()),
		revealing:         collections.NewKeySet(sb, types.RevealingKeyPrefix, "revealing", collcodec.NewBytesKey[types.DataRequestIndex]()),
		tallying:          collections.NewKeySet(sb, types.TallyingKeyPrefix, "tallying", collcodec.NewBytesKey[types.DataRequestIndex]()),
		timeoutQueue:      collections.NewKeySet(sb, types.TimeoutQueueKeyPrefix, "timeout_queue", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) UpdateDataRequestTimeout(ctx sdk.Context, drID string, oldTimeoutHeight, newTimeoutHeight uint64) error {
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

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

func (k Keeper) GetDataRequestConfig(ctx sdk.Context) (types.DataRequestConfig, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return types.DataRequestConfig{}, err
	}
	return params.DataRequestConfig, nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
