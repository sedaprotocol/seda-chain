package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
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

	Schema             collections.Schema
	currentBatchNumber collections.Sequence
	batch              collections.Map[uint64, types.Batch]
	params             collections.Item[types.Params]
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
		currentBatchNumber:    collections.NewSequence(sb, types.CurrentBatchNumberKey, "current_batch_number"),
		batch:                 collections.NewMap(sb, types.BatchPrefix, "batch", collections.Uint64Key, codec.CollValue[types.Batch](cdc)),
		params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

func (k Keeper) SetCurrentBatchNum(ctx context.Context, batchNum uint64) error {
	return k.currentBatchNumber.Set(ctx, batchNum)
}

func (k Keeper) IncrementCurrentBatchNum(ctx context.Context) error {
	_, err := k.currentBatchNumber.Next(ctx)
	return err
}

func (k Keeper) GetCurrentBatchNum(ctx context.Context) (uint64, error) {
	batchNum, err := k.currentBatchNumber.Peek(ctx)
	if err != nil {
		return 0, err
	}
	return batchNum, nil
}

func (k Keeper) SetBatch(ctx context.Context, batch types.Batch) error {
	return k.batch.Set(ctx, batch.BatchNumber, batch)
}

func (k Keeper) GetBatch(ctx context.Context, batchNum uint64) (types.Batch, error) {
	batch, err := k.batch.Get(ctx, batchNum)
	if err != nil {
		return types.Batch{}, err
	}
	return batch, nil
}

// GetPreviousDataResultRoot returns the previous batch's data result
// tree root in byte slice. If there is no previous batch, it returns
// an empty byte slice.
func (k Keeper) GetPreviousDataResultRoot(ctx context.Context) ([]byte, error) {
	curBatchNum, err := k.GetCurrentBatchNum(ctx)
	if err != nil {
		return nil, err
	}
	if curBatchNum == 0 {
		return []byte{}, err
	}
	batch, err := k.batch.Get(ctx, curBatchNum-1)
	if err != nil {
		return nil, err
	}
	root, err := hex.DecodeString(batch.DataResultRoot)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// IterateBatches iterates over the batches and performs a given
// callback function.
func (k Keeper) IterateBatches(ctx sdk.Context, callback func(types.Batch) (stop bool)) error {
	iter, err := k.batch.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if callback(kv.Value) {
			break
		}
	}
	return nil
}

// GetAllBatches returns all batches in the store.
func (k Keeper) GetAllBatches(ctx sdk.Context) ([]types.Batch, error) {
	var batches []types.Batch
	err := k.IterateBatches(ctx, func(batch types.Batch) bool {
		batches = append(batches, batch)
		return false
	})
	if err != nil {
		return nil, err
	}
	return batches, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
