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
	stakingKeeper         types.StakingKeeper
	wasmStorageKeeper     types.WasmStorageKeeper
	pubKeyKeeper          types.PubKeyKeeper
	wasmKeeper            wasmtypes.ContractOpsKeeper
	wasmViewKeeper        wasmtypes.ViewKeeper
	validatorAddressCodec addresscodec.Codec

	// authority is the address capable of executing MsgUpdateParams.
	// Typically, this should be the gov module address.
	authority string

	Schema                collections.Schema
	dataResults           collections.Map[collections.Pair[bool, string], types.DataResult]
	batchAssignments      collections.Map[string, uint64]
	currentBatchNumber    collections.Sequence
	batches               *collections.IndexedMap[int64, types.Batch, BatchIndexes]
	validatorTreeEntries  collections.Map[collections.Pair[uint64, []byte], types.ValidatorTreeEntry]
	dataResultTreeEntries collections.Map[uint64, types.DataResultTreeEntries]
	batchSignatures       collections.Map[collections.Pair[uint64, []byte], types.BatchSignatures]
	params                collections.Item[types.Params]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	authority string,
	sk types.StakingKeeper,
	wsk types.WasmStorageKeeper,
	pkk types.PubKeyKeeper,
	wk wasmtypes.ContractOpsKeeper,
	wvk wasmtypes.ViewKeeper,
	validatorAddressCodec addresscodec.Codec,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		stakingKeeper:         sk,
		wasmStorageKeeper:     wsk,
		pubKeyKeeper:          pkk,
		wasmKeeper:            wk,
		wasmViewKeeper:        wvk,
		validatorAddressCodec: validatorAddressCodec,
		authority:             authority,
		dataResults:           collections.NewMap(sb, types.DataResultsPrefix, "data_results", collections.PairKeyCodec(collections.BoolKey, collections.StringKey), codec.CollValue[types.DataResult](cdc)),
		batchAssignments:      collections.NewMap(sb, types.BatchAssignmentsPrefix, "batch_assignments", collections.StringKey, collections.Uint64Value),
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

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
