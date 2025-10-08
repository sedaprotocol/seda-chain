package keeper

import (
	"encoding/hex"
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
	// commits is a map of data request ID - public key pairs to commits.
	commits collections.Map[collections.Pair[string, string], []byte]
	// revealers is a set of data request ID - public key pairs.
	revealers collections.KeySet[collections.Pair[string, string]]
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
		commits:           collections.NewMap(sb, types.CommitsPrefix, "commits", collections.PairKeyCodec(collections.StringKey, collections.StringKey), collections.BytesValue),
		revealers:         collections.NewKeySet(sb, types.RevealersPrefix, "revealers", collections.PairKeyCodec(collections.StringKey, collections.StringKey)),
		revealBodies:      collections.NewMap(sb, types.RevealBodiesKeyPrefix, "reveal_bodies", collections.PairKeyCodec(collections.StringKey, collections.StringKey), codec.CollValue[types.RevealBody](cdc)),
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
func (k Keeper) LoadRevealsHashSorted(ctx sdk.Context, drID string, revealers []string, entropy []byte) ([]types.Reveal, []string, []uint64) {
	reveals := make([]types.Reveal, len(revealers))
	i := 0
	for _, revealer := range revealers {
		revealBody, err := k.GetRevealBody(ctx, drID, revealer)
		if err != nil {
			// TODO Proper error handling
			panic(err)
		}
		reveals[i] = types.Reveal{Executor: revealer, RevealBody: revealBody}
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

// GetRevealBodies returns a map of hex-encoded executor public keys to their
// reveal bodies for the given data request ID.
func (k Keeper) GetRevealBodies(ctx sdk.Context, drID string) (map[string]*types.RevealBody, error) {
	rng := collections.NewPrefixedPairRange[string, string](drID)
	iter, err := k.revealBodies.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	revealBodies := make(map[string]*types.RevealBody)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		value, err := iter.Value()
		if err != nil {
			return nil, err
		}
		revealBodies[key.K2()] = &value
	}
	return revealBodies, nil
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

// AddCommit adds a commit to the given data request ID and public key pair and
// returns the updated count of commits. It return an error if the commit already exists.
func (k Keeper) AddCommit(ctx sdk.Context, drID, publicKey string, commit []byte) (uint32, error) {
	ranger := collections.NewPrefixedPairRange[string, string](drID)
	var count uint32
	err := k.commits.Walk(ctx, ranger, func(key collections.Pair[string, string], value []byte) (stop bool, err error) {
		if key.K2() == publicKey {
			return true, types.ErrAlreadyCommitted
		}
		count++
		return false, nil
	})
	if err != nil {
		return 0, err
	}

	err = k.commits.Set(ctx, collections.Join(drID, publicKey), commit)
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

// GetCommit returns a commit given a data request ID - public key pair.
func (k Keeper) GetCommit(ctx sdk.Context, drID, publicKey string) ([]byte, error) {
	return k.commits.Get(ctx, collections.Join(drID, publicKey))
}

// GetCommits returns a map of hex-encoded public keys to their commits for the
// given data request ID.
func (k Keeper) GetCommits(ctx sdk.Context, drID string) (map[string]string, error) {
	rng := collections.NewPrefixedPairRange[string, string](drID)
	iter, err := k.commits.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	commits := make(map[string]string)
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		value, err := iter.Value()
		if err != nil {
			return nil, err
		}
		commits[key.K2()] = hex.EncodeToString(value)
	}
	return commits, nil
}

// GetCommitters returns the list of hex-encoded executor public keys that have
// committed for the given data request.
func (k Keeper) GetCommitters(ctx sdk.Context, drID string) ([]string, error) {
	rng := collections.NewPrefixedPairRange[string, string](drID)
	iter, err := k.commits.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var committers []string
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		committers = append(committers, key.K2())
	}
	return committers, nil
}

// HasCommitted checks if the executor with the given public key has committed
// for the given data request.
func (k Keeper) HasCommitted(ctx sdk.Context, drID, publicKey string) (bool, error) {
	return k.commits.Has(ctx, collections.Join(drID, publicKey))
}

// RemoveCommits removes all commits stored for the given data request.
func (k Keeper) RemoveCommits(ctx sdk.Context, drID string) error {
	iter, err := k.commits.Iterate(ctx, collections.NewPrefixedPairRange[string, string](drID))
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		err = k.commits.Remove(ctx, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// MarkAsRevealed adds the given public key to the list of the given data request's
// revealers and returns an updated count of reveals. It returns an error if the
// public key already exists.
func (k Keeper) MarkAsRevealed(ctx sdk.Context, drID, publicKey string) (uint32, error) {
	ranger := collections.NewPrefixedPairRange[string, string](drID)
	var count uint32
	err := k.revealers.Walk(ctx, ranger, func(key collections.Pair[string, string]) (stop bool, err error) {
		if key.K2() == publicKey {
			return true, types.ErrAlreadyRevealed
		}
		count++
		return false, nil
	})
	if err != nil {
		return 0, err
	}

	err = k.revealers.Set(ctx, collections.Join(drID, publicKey))
	if err != nil {
		return 0, err
	}
	return count + 1, nil
}

// HasRevealed checks if the executor with the given public key has revealed for
// the given data request.
func (k Keeper) HasRevealed(ctx sdk.Context, drID, publicKey string) (bool, error) {
	return k.revealers.Has(ctx, collections.Join(drID, publicKey))
}

// GetRevealers returns the list of hex-encoded executor public keys that have
// revealed for the given data request.
func (k Keeper) GetRevealers(ctx sdk.Context, drID string) ([]string, error) {
	rng := collections.NewPrefixedPairRange[string, string](drID)
	iter, err := k.revealers.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var revealers []string
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, err
		}
		revealers = append(revealers, key.K2())
	}
	return revealers, nil
}

// RemoveRevealers removes all revealers stored for the given data request.
func (k Keeper) RemoveRevealers(ctx sdk.Context, drID string) error {
	iter, err := k.revealers.Iterate(ctx, collections.NewPrefixedPairRange[string, string](drID))
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return err
		}
		err = k.revealers.Remove(ctx, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCommittersAndRevealers returns the list of hex-encoded executor public keys
// that have committed and revealed for the given data request.
func (k Keeper) GetCommittersAndRevealers(ctx sdk.Context, drID string) ([]string, []string, error) {
	committers, err := k.GetCommitters(ctx, drID)
	if err != nil {
		return nil, nil, err
	}
	if len(committers) == 0 {
		committers = []string{}
	}

	revealers, err := k.GetRevealers(ctx, drID)
	if err != nil {
		return nil, nil, err
	}
	if len(revealers) == 0 {
		revealers = []string{}
	}
	return committers, revealers, nil
}

// ClearDataRequest removes all objects corresponding to the given data request
// from the core module store.
func (k Keeper) ClearDataRequest(ctx sdk.Context, index types.DataRequestIndex, status types.DataRequestStatus) error {
	err := k.RemoveRevealBodies(ctx, index.DrID())
	if err != nil {
		return err
	}
	err = k.RemoveDataRequest(ctx, index, status)
	if err != nil {
		return err
	}
	err = k.RemoveCommits(ctx, index.DrID())
	if err != nil {
		return err
	}
	err = k.RemoveRevealers(ctx, index.DrID())
	if err != nil {
		return err
	}
	return nil
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
