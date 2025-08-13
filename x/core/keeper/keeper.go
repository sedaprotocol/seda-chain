package keeper

import (
	"fmt"
	"sort"

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
	dataProxyKeeper   types.DataProxyKeeper
	stakingKeeper     types.StakingKeeper
	bankKeeper        types.BankKeeper
	wasmKeeper        wasmtypes.ContractOpsKeeper
	wasmViewKeeper    wasmtypes.ViewKeeper
	authority         string

	Schema    collections.Schema
	Allowlist collections.KeySet[string]
	Stakers   collections.Map[string, types.Staker]
	params    collections.Item[types.Params]

	dataRequests collections.Map[string, types.DataRequest]
	revealBodies collections.Map[collections.Pair[string, string], types.RevealBody]
	committing   collections.KeySet[types.DataRequestIndex]
	revealing    collections.KeySet[types.DataRequestIndex]
	tallying     collections.KeySet[types.DataRequestIndex]
	timeoutQueue collections.KeySet[collections.Pair[uint64, string]]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	wsk types.WasmStorageKeeper,
	batk types.BatchingKeeper,
	dpk types.DataProxyKeeper,
	sk types.StakingKeeper,
	bank types.BankKeeper,
	wk wasmtypes.ContractOpsKeeper,
	wvk wasmtypes.ViewKeeper,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		wasmStorageKeeper: wsk,
		batchingKeeper:    batk,
		dataProxyKeeper:   dpk,
		stakingKeeper:     sk,
		bankKeeper:        bank,
		wasmKeeper:        wk,
		wasmViewKeeper:    wvk,
		authority:         authority,
		Allowlist:         collections.NewKeySet(sb, types.AllowlistKey, "allowlist", collections.StringKey),
		Stakers:           collections.NewMap(sb, types.StakersKeyPrefix, "stakers", collections.StringKey, codec.CollValue[types.Staker](cdc)),
		params:            collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		dataRequests:      collections.NewMap(sb, types.DataRequestsKeyPrefix, "data_requests", collections.StringKey, codec.CollValue[types.DataRequest](cdc)),
		revealBodies:      collections.NewMap(sb, types.RevealsKeyPrefix, "reveals", collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollValue[types.RevealBody](cdc)),
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

// GetDataRequest retrieves a data request given its hex-encoded ID.
func (k Keeper) GetDataRequest(ctx sdk.Context, id string) (types.DataRequest, error) {
	return k.dataRequests.Get(ctx, id)
}

// HasDataRequest checks if a data request exists given its hex-encoded ID.
func (k Keeper) HasDataRequest(ctx sdk.Context, id string) (bool, error) {
	return k.dataRequests.Has(ctx, id)
}

// SetDataRequest stores a data request in the store.
func (k Keeper) SetDataRequest(ctx sdk.Context, dr types.DataRequest) error {
	return k.dataRequests.Set(ctx, dr.Id, dr)
}

// RemoveDataRequest removes a data request given its hex-encoded ID.
func (k Keeper) RemoveDataRequest(ctx sdk.Context, id string) error {
	return k.dataRequests.Remove(ctx, id)
}

func (k Keeper) GetRevealBody(ctx sdk.Context, drID string, executor string) (types.RevealBody, error) {
	return k.revealBodies.Get(ctx, collections.Join(drID, executor))
}

func (k Keeper) SetRevealBody(ctx sdk.Context, drID string, executor string, revealBody types.RevealBody) error {
	return k.revealBodies.Set(ctx, collections.Join(drID, executor), revealBody)
}

// RemoveRevealBodies removes reveal bodies corresponding to a given data request.
func (k Keeper) RemoveRevealBodies(ctx sdk.Context, drID string) error {
	iter, err := k.revealBodies.Iterate(ctx, collections.NewPrefixedPairRange[string, string](drID))
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		k.revealBodies.Remove(ctx, collections.Join(drID, key.K2()))
	}
	return nil
}

// LoadRevealsSorted returns reveals, executors, and gas reports sorted in a
// deterministically random order. The reveals are retrieved based on the given
// map of executors, and each reveal's reported proxy public keys are sorted.
func (k Keeper) LoadRevealsSorted(ctx sdk.Context, drID string, revealsMap map[string]bool) ([]types.Reveal, []string, []uint64) {
	reveals := make([]types.Reveal, len(revealsMap))
	i := 0
	for executor := range revealsMap {
		revealBody, err := k.GetRevealBody(ctx, drID, executor)
		if err != nil {
			// TODO Proper error handling
			return nil, nil, nil
		}
		reveals[i] = types.Reveal{Executor: executor, RevealBody: revealBody}
		sort.Strings(reveals[i].ProxyPubKeys)
		i++
	}

	sortedReveals := types.HashSort(reveals, types.GetEntropy(drID, ctx.BlockHeight()))

	executors := make([]string, len(sortedReveals))
	gasReports := make([]uint64, len(sortedReveals))
	for i, reveal := range sortedReveals {
		executors[i] = reveal.Executor
		gasReports[i] = reveal.GasUsed
	}
	return sortedReveals, executors, gasReports
}

func (k Keeper) RemoveFromTimeoutQueue(ctx sdk.Context, drID string, timeoutHeight uint64) error {
	err := k.timeoutQueue.Remove(ctx, collections.Join(timeoutHeight, drID))
	if err != nil {
		return err
	}
	return nil
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
	params, err := k.params.Get(ctx)
	if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetDataRequestConfig(ctx sdk.Context) (types.DataRequestConfig, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.DataRequestConfig{}, err
	}
	return params.DataRequestConfig, nil
}

func (k Keeper) GetStakingConfig(ctx sdk.Context) (types.StakingConfig, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.StakingConfig{}, err
	}
	return params.StakingConfig, nil
}

func (k Keeper) GetTallyConfig(ctx sdk.Context) (types.TallyConfig, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.TallyConfig{}, err
	}
	return params.TallyConfig, nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
