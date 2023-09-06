package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// SetDataRequestWasm stores Data Request Wasm using its hash as the key.
func (k Keeper) SetDataRequestWasm(ctx sdk.Context, wasm *types.Wasm) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalLengthPrefixed(wasm)
	store.Set(types.GetDataRequestWasmKey(wasm.Hash), bz)
}

// GetDataRequestWasm returns Data Request Wasm given its key.
func (k Keeper) GetDataRequestWasm(ctx sdk.Context, hash []byte) *types.Wasm {
	var wasm *types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDataRequestWasmKey(hash))
	k.cdc.MustUnmarshalLengthPrefixed(bz, wasm)
	return wasm
}

// SetOverlayWasm stores Overlay Wasm using its hash as the key.
func (k Keeper) SetOverlayWasm(ctx sdk.Context, wasm *types.Wasm) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalLengthPrefixed(wasm)
	store.Set(types.GetDataRequestWasmKey(wasm.Hash), bz)
}

// GetOverlayWasm returns Overlay Wasm given its key.
func (k Keeper) GetOverlayWasm(ctx sdk.Context, hash []byte) *types.Wasm {
	var wasm *types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOverlayWasmKey(hash))
	k.cdc.MustUnmarshalLengthPrefixed(bz, wasm)
	return wasm
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
