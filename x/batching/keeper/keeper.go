package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

type Keeper struct {
	stakingKeeper types.StakingKeeper

	// authority is the address capable of executing MsgUpdateParams.
	// Typically, this should be the gov module address.
	authority string

	Schema             collections.Schema
	currentBatchNumber collections.Sequence
	batch              collections.Map[uint64, types.Batch]
	params             collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string, sk types.StakingKeeper) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		authority:          authority,
		stakingKeeper:      sk,
		currentBatchNumber: collections.NewSequence(sb, types.CurrentBatchNumberKey, "current_batch_number"),
		batch:              collections.NewMap(sb, types.BatchPrefix, "batch", collections.Uint64Key, codec.CollValue[types.Batch](cdc)),
		params:             collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
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

func (k Keeper) IncrementCurrentBatchNum(ctx context.Context, batchNum uint64) error {
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

func (k Keeper) SetBatch(ctx context.Context, batchNum uint64, batch types.Batch) error {
	return k.batch.Set(ctx, batchNum, batch)
}

func (k Keeper) GetBatch(ctx context.Context, batchNum uint64) (types.Batch, error) {
	batch, err := k.batch.Get(ctx, batchNum)
	if err != nil {
		return types.Batch{}, err
	}
	return batch, nil
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
