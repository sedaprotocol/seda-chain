package keeper

import (
	"fmt"
	"sort"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	collcdc "cosmossdk.io/collections/codec"
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

	Schema collections.Schema

	// Staking-related states:
	// allowlist is an owner-controlled allowlist of staker public keys.
	allowlist collections.KeySet[string]
	// stakers is a map of staker public keys to staker objects.
	stakers collections.Map[string, types.Staker]

	// Data request-related states:
	// dataRequests is a map of data request IDs to data request objects.
	dataRequests collections.Map[string, types.DataRequest]
	// revealBodies is a map of data request IDs and executor public keys to reveal bodies.
	revealBodies collections.Map[collections.Pair[string, string], types.RevealBody]
	// dataRequestIndexing is a set of data request indices under different statuses.
	dataRequestIndexing collections.KeySet[collections.Pair[types.DataRequestStatus, types.DataRequestIndex]]
	// timeoutQueue is a queue of data request IDs and their timeout heights.
	timeoutQueue collections.KeySet[collections.Pair[int64, string]]

	// Parameter state:
	// params defines the core module parameters.
	params collections.Item[types.Params]
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
		wasmStorageKeeper:   wsk,
		batchingKeeper:      batk,
		dataProxyKeeper:     dpk,
		stakingKeeper:       sk,
		bankKeeper:          bank,
		wasmKeeper:          wk,
		wasmViewKeeper:      wvk,
		authority:           authority,
		allowlist:           collections.NewKeySet(sb, types.AllowlistKey, "allowlist", collections.StringKey),
		stakers:             collections.NewMap(sb, types.StakersKeyPrefix, "stakers", collections.StringKey, codec.CollValue[types.Staker](cdc)),
		dataRequests:        collections.NewMap(sb, types.DataRequestsKeyPrefix, "data_requests", collections.StringKey, codec.CollValue[types.DataRequest](cdc)),
		revealBodies:        collections.NewMap(sb, types.RevealBodiesKeyPrefix, "reveals", collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollValue[types.RevealBody](cdc)),
		dataRequestIndexing: collections.NewKeySet(sb, types.DrIndexingKeyPrefix, "data_request_indexing", collections.PairKeyCodec(collcdc.NewInt32Key[types.DataRequestStatus](), collcdc.NewBytesKey[types.DataRequestIndex]())),
		timeoutQueue:        collections.NewKeySet(sb, types.TimeoutQueueKeyPrefix, "timeout_queue", collections.PairKeyCodec(collections.Int64Key, collections.StringKey)),
		params:              collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
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
	return k.dataRequests.Set(ctx, dr.ID, dr)
}

// RemoveDataRequest removes a data request given its hex-encoded ID.
func (k Keeper) RemoveDataRequest(ctx sdk.Context, id string) error {
	return k.dataRequests.Remove(ctx, id)
}

// GetRevealBody retrieves a reveal body given a data request ID and an executor.
func (k Keeper) GetRevealBody(ctx sdk.Context, drID string, executor string) (types.RevealBody, error) {
	return k.revealBodies.Get(ctx, collections.Join(drID, executor))
}

// SetRevealBody stores a reveal body in the store.
func (k Keeper) SetRevealBody(ctx sdk.Context, executor string, revealBody types.RevealBody) error {
	return k.revealBodies.Set(ctx, collections.Join(revealBody.DrId, executor), revealBody)
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
		err = k.revealBodies.Remove(ctx, collections.Join(drID, key.K2()))
		if err != nil {
			return err
		}
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

// GetParams retrieves the core module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

// SetParams stores the core module parameters.
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
