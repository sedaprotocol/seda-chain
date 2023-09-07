package keeper

import (
	"encoding/hex"
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
	bz := k.cdc.MustMarshal(wasm)
	store.Set(types.GetDataRequestWasmKey(wasm.Hash), bz)
}

// GetDataRequestWasm returns Data Request Wasm given its key.
func (k Keeper) GetDataRequestWasm(ctx sdk.Context, hash []byte) *types.Wasm {
	var wasm *types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDataRequestWasmKey(hash))
	k.cdc.MustUnmarshal(bz, wasm)
	return wasm
}

// IterateAllDataRequestWasms iterates over the all the stored Data Request
// Wasms and performs a given callback function.
func (k Keeper) IterateAllDataRequestWasms(ctx sdk.Context, callback func(wasm types.Wasm) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixDataRequest)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var wasm types.Wasm
		k.cdc.MustUnmarshal(iterator.Value(), &wasm)

		if callback(wasm) {
			break
		}
	}
}

// GetDataRequestWasmHashes returns hashes of all Data Request Wasms
// in the store.
func (k Keeper) GetDataRequestWasmHashes(ctx sdk.Context) []string {
	var hashes []string
	k.IterateAllDataRequestWasms(ctx, func(w types.Wasm) bool {
		hashes = append(hashes, hex.EncodeToString(w.Hash))
		return false
	})
	return hashes
}

// SetOverlayWasm stores Overlay Wasm using its hash as the key.
func (k Keeper) SetOverlayWasm(ctx sdk.Context, wasm *types.Wasm) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(wasm)
	store.Set(types.GetDataRequestWasmKey(wasm.Hash), bz)
}

// GetOverlayWasm returns Overlay Wasm given its key.
func (k Keeper) GetOverlayWasm(ctx sdk.Context, hash []byte) *types.Wasm {
	var wasm *types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOverlayWasmKey(hash))
	k.cdc.MustUnmarshal(bz, wasm)
	return wasm
}

// IterateAllOverlayWasms iterates over the all the stored Overlay Wasms
// and performs a given callback function.
func (k Keeper) IterateAllOverlayWasms(ctx sdk.Context, callback func(wasm types.Wasm) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOverlay)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var wasm types.Wasm
		k.cdc.MustUnmarshal(iterator.Value(), &wasm)

		if callback(wasm) {
			break
		}
	}
}

// GetOverlayWasmHashes returns hashes of all Overlay Wasms in the store.
func (k Keeper) GetOverlayWasmHashes(ctx sdk.Context) []string {
	var hashes []string
	k.IterateAllOverlayWasms(ctx, func(w types.Wasm) bool {
		hashes = append(hashes, hex.EncodeToString(w.Hash))
		return false
	})
	return hashes
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
