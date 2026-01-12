package keeper

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	addresscodec "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

type Keeper struct {
	// authority is the address capable of executing MsgUpdateParams.
	// Typically, this should be the gov module address.
	authority string

	stakingKeeper         types.StakingKeeper
	slashingKeeper        types.SlashingKeeper
	wasmStorageKeeper     types.WasmStorageKeeper
	pubKeyKeeper          types.PubKeyKeeper
	wasmKeeper            wasmtypes.ContractOpsKeeper
	wasmViewKeeper        wasmtypes.ViewKeeper
	validatorAddressCodec addresscodec.Codec

	Schema collections.Schema

	batchAssignments      collections.Map[collections.Pair[string, uint64], uint64]
	currentBatchNumber    collections.Sequence
	batches               *collections.IndexedMap[int64, types.Batch, BatchIndexes]
	validatorTreeEntries  collections.Map[collections.Pair[uint64, []byte], types.ValidatorTreeEntry]
	dataResultTreeEntries collections.Map[uint64, types.DataResultTreeEntries]
	batchSignatures       collections.Map[collections.Pair[uint64, []byte], types.BatchSignatures]
	params                collections.Item[types.Params]
	// dataResults is the newer version of dataResults. The items in this collection
	// have corresponding items in batchDataResults.
	dataResults collections.Map[collections.Triple[bool, string, uint64], types.DataResult]
	// batchDataResults maps batch number to a list of data request ID - posted height
	// pairs to support basic pruning of data results.
	batchDataResults collections.Map[uint64, types.DataRequestIDHeights]
	// legacyDataResults is the older version of dataResults. The items in this
	// collection do not have corresponding items in batchDataResults.
	legacyDataResults collections.Map[collections.Triple[bool, string, uint64], types.DataResult]
	// hasPruningCaughtUp is switched to true when either of the following conditions
	// is met:
	// (i) All batches up to batchNumberAtUpgrade have been pruned.
	// (ii) All batches up to (currentBatchNum - numBatchesToKeep) have been pruned.
	hasPruningCaughtUp collections.Item[bool]
	// batchNumberAtUpgrade is the batch number of the latest batch at upgrade time
	// except when its value is 0, in which case there was no upgrade.
	batchNumberAtUpgrade collections.Item[uint64]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	sk types.StakingKeeper,
	slk types.SlashingKeeper,
	wsk types.WasmStorageKeeper,
	pkk types.PubKeyKeeper,
	wk wasmtypes.ContractOpsKeeper,
	wvk wasmtypes.ViewKeeper,
	validatorAddressCodec addresscodec.Codec,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		authority:             authority,
		stakingKeeper:         sk,
		slashingKeeper:        slk,
		wasmStorageKeeper:     wsk,
		pubKeyKeeper:          pkk,
		wasmKeeper:            wk,
		wasmViewKeeper:        wvk,
		validatorAddressCodec: validatorAddressCodec,
		legacyDataResults:     collections.NewMap(sb, types.LegacyDataResultsPrefix, "legacy_data_results", collections.TripleKeyCodec(collections.BoolKey, collections.StringKey, collections.Uint64Key), codec.CollValue[types.DataResult](cdc)),
		hasPruningCaughtUp:    collections.NewItem(sb, types.HasPruningCaughtUpKey, "has_pruning_caught_up", collections.BoolValue),
		dataResults:           collections.NewMap(sb, types.DataResultsPrefix, "data_results", collections.TripleKeyCodec(collections.BoolKey, collections.StringKey, collections.Uint64Key), codec.CollValue[types.DataResult](cdc)),
		batchAssignments:      collections.NewMap(sb, types.BatchAssignmentsPrefix, "batch_assignments", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key), collections.Uint64Value),
		batchDataResults:      collections.NewMap(sb, types.BatchDataResultsPrefix, "batch_data_results", collections.Uint64Key, codec.CollValue[types.DataRequestIDHeights](cdc)),
		batchNumberAtUpgrade:  collections.NewItem(sb, types.BatchNumberAtUpgradeKey, "batch_number_at_upgrade", collections.Uint64Value),
		currentBatchNumber:    collections.NewSequence(sb, types.CurrentBatchNumberKey, "current_batch_number"),
		batches:               collections.NewIndexedMap(sb, types.BatchesKeyPrefix, "batches", collections.Int64Key, codec.CollValue[types.Batch](cdc), NewBatchIndexes(sb)),
		validatorTreeEntries:  collections.NewMap(sb, types.ValidatorTreeEntriesKeyPrefix, "validator_tree_entries", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), codec.CollValue[types.ValidatorTreeEntry](cdc)),
		dataResultTreeEntries: collections.NewMap(sb, types.DataResultTreeEntriesKeyPrefix, "data_result_tree_entries", collections.Uint64Key, codec.CollValue[types.DataResultTreeEntries](cdc)),
		batchSignatures:       collections.NewMap(sb, types.BatchSignaturesKeyPrefix, "batch_signatures", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), codec.CollValue[types.BatchSignatures](cdc)),
		params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

func NewBatchIndexes(sb *collections.SchemaBuilder) BatchIndexes {
	return BatchIndexes{
		Number: indexes.NewUnique(
			sb, types.BatchNumberKeyPrefix, "batch_by_number", collections.Uint64Key, collections.Int64Key,
			func(_ int64, batch types.Batch) (uint64, error) {
				return batch.BatchNumber, nil
			},
		),
	}
}

type BatchIndexes struct {
	// Number is a unique index that indexes batches by their batch number.
	Number *indexes.Unique[uint64, int64, types.Batch]
}

func (i BatchIndexes) IndexesList() []collections.Index[int64, types.Batch] {
	return []collections.Index[int64, types.Batch]{
		i.Number,
	}
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
