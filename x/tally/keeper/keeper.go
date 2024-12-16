package keeper

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

type Keeper struct {
	wasmStorageKeeper types.WasmStorageKeeper
	batchingKeeper    types.BatchingKeeper
	wasmKeeper        wasmtypes.ContractOpsKeeper
	wasmViewKeeper    wasmtypes.ViewKeeper
	authority         string

	Schema collections.Schema
	params collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, wsk types.WasmStorageKeeper, bk types.BatchingKeeper, wk wasmtypes.ContractOpsKeeper, wvk wasmtypes.ViewKeeper, authority string) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		wasmStorageKeeper: wsk,
		batchingKeeper:    bk,
		wasmKeeper:        wk,
		wasmViewKeeper:    wvk,
		params:            collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
		authority:         authority,
	}
	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) GetMaxTallyGasLimit(ctx sdk.Context) (uint64, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return 0, err
	}
	return params.MaxTallyGasLimit, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
