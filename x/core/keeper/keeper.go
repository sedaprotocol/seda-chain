package keeper

import (
	"fmt"
	"sort"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
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

	feeCollectorName string
	txDecoder        sdk.TxDecoder

	Schema collections.Schema

	// Core module parameter states:
	owner        collections.Item[string]
	pendingOwner collections.Item[string]
	paused       collections.Item[bool]
	params       collections.Item[types.Params]

	// Staking-related states:
	// allowlist is an owner-controlled allowlist of staker public keys.
	allowlist collections.KeySet[string]
	// stakers is a map of staker public keys to staker objects.
	stakers types.StakerIndexing

	// Data request-related states:
	// dataRequestIndexing is a set of data request indices under different statuses.
	dataRequests types.DataRequestIndexing
	// revealBodies is a map of data request IDs and executor public keys to reveal bodies.
	revealBodies collections.Map[collections.Pair[string, string], types.RevealBody]
	// timeoutQueue is a queue of data request IDs and their timeout heights.
	timeoutQueue collections.KeySet[collections.Pair[int64, string]]
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
	feeCollectorName string,
	txDecoder sdk.TxDecoder,
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
		feeCollectorName:  feeCollectorName,
		txDecoder:         txDecoder,
		owner:             collections.NewItem(sb, types.OwnerKey, "owner", collections.StringValue),
		paused:            collections.NewItem(sb, types.PausedKey, "paused", collections.BoolValue),
		pendingOwner:      collections.NewItem(sb, types.PendingOwnerKey, "pending_owner", collections.StringValue),
		allowlist:         collections.NewKeySet(sb, types.AllowlistKey, "allowlist", collections.StringKey),
		stakers:           types.NewStakerIndexing(sb, cdc),
		dataRequests:      types.NewDataRequestIndexing(sb, cdc),
		revealBodies:      collections.NewMap(sb, types.RevealBodiesKeyPrefix, "reveals", collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollValue[types.RevealBody](cdc)),
		timeoutQueue:      collections.NewKeySet(sb, types.TimeoutQueueKeyPrefix, "timeout_queue", collections.PairKeyCodec(collections.Int64Key, collections.StringKey)),
		params:            collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// LoadRevealsHashSorted returns reveals, executors, and gas reports corresponding
// to the given data request ID and map of reveal executors' public keys. Returned
// slices are in the same order sorted by the hash of the executor public key and
// given entropy. If the entropy is nil, the items are sorted simply by the executor
// public key without hashing.
func (k Keeper) LoadRevealsHashSorted(ctx sdk.Context, drID string, revealsMap map[string]bool, entropy []byte) ([]types.Reveal, []string, []uint64) {
	reveals := make([]types.Reveal, len(revealsMap))
	i := 0
	for executor := range revealsMap {
		revealBody, err := k.GetRevealBody(ctx, drID, executor)
		if err != nil {
			// TODO Proper error handling
			panic(err)
		}
		reveals[i] = types.Reveal{Executor: executor, RevealBody: revealBody}
		sort.Strings(reveals[i].ProxyPubKeys)
		i++
	}

	sortedReveals := types.HashSort(reveals, entropy)

	executors := make([]string, len(sortedReveals))
	gasReports := make([]uint64, len(sortedReveals))
	for i, reveal := range sortedReveals {
		executors[i] = reveal.Executor
		gasReports[i] = reveal.GasUsed
	}
	return sortedReveals, executors, gasReports
}

// GetRevealBody retrieves a reveal body given a data request ID and an executor.
func (k Keeper) GetRevealBody(ctx sdk.Context, drID string, executor string) (types.RevealBody, error) {
	return k.revealBodies.Get(ctx, collections.Join(drID, executor))
}

// SetRevealBody stores a reveal body in the store given the hex-encoded ID of
// the executor.
func (k Keeper) SetRevealBody(ctx sdk.Context, executor string, revealBody types.RevealBody) error {
	return k.revealBodies.Set(ctx, collections.Join(revealBody.DrID, executor), revealBody)
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

// GetAllRevealBodies retrieves all reveal bodies from the store.
func (k Keeper) GetAllRevealBodies(ctx sdk.Context) ([]types.RevealBody, error) {
	iter, err := k.revealBodies.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	revealBodies := make([]types.RevealBody, 0)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		revealBody, err := k.GetRevealBody(ctx, key.K1(), key.K2())
		if err != nil {
			return nil, err
		}
		revealBodies = append(revealBodies, revealBody)
	}
	return revealBodies, nil
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
	return *params.DataRequestConfig, nil
}

func (k Keeper) GetStakingConfig(ctx sdk.Context) (types.StakingConfig, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.StakingConfig{}, err
	}
	return *params.StakingConfig, nil
}

func (k Keeper) GetTallyConfig(ctx sdk.Context) (types.TallyConfig, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return types.TallyConfig{}, err
	}
	return *params.TallyConfig, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
