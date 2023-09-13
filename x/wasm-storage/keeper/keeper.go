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
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string) *Keeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
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

// HasDataRequestWasm checks if a given Data Request Wasm exists.
func (k Keeper) HasDataRequestWasm(ctx sdk.Context, wasm *types.Wasm) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetDataRequestWasmKey(wasm.Hash))
}

// SetOverlayWasm stores Overlay Wasm using its hash as the key.
func (k Keeper) SetOverlayWasm(ctx sdk.Context, wasm *types.Wasm) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(wasm)
	store.Set(types.GetOverlayWasmKey(wasm.Hash), bz)
}

// GetOverlayWasm returns Overlay Wasm given its key.
func (k Keeper) GetOverlayWasm(ctx sdk.Context, hash []byte) *types.Wasm {
	var wasm *types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOverlayWasmKey(hash))
	k.cdc.MustUnmarshal(bz, wasm)
	return wasm
}

// HasOverlayWasm checks if a given Overlay Wasm exists.
func (k Keeper) HasOverlayWasm(ctx sdk.Context, wasm *types.Wasm) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetOverlayWasmKey(wasm.Hash))
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

// ListDataRequestWasms returns hashes and types of all Data Request Wasms
// in the store.
func (k Keeper) ListDataRequestWasms(ctx sdk.Context) []string {
	var hashTypePairs []string
	k.IterateAllDataRequestWasms(ctx, func(w types.Wasm) bool {
		hashTypePairs = append(hashTypePairs, hex.EncodeToString(w.Hash)+","+w.WasmType.String())
		return false
	})
	return hashTypePairs
}

// ListOverlayWasms returns hashes and types of all Overlay Wasms in the store.
func (k Keeper) ListOverlayWasms(ctx sdk.Context) []string {
	var hashTypePairs []string
	k.IterateAllOverlayWasms(ctx, func(w types.Wasm) bool {
		hashTypePairs = append(hashTypePairs, hex.EncodeToString(w.Hash)+","+w.WasmType.String())
		return false
	})
	return hashTypePairs
}

func (k Keeper) GetAllWasms(ctx sdk.Context) []types.Wasm {
	var wasms []types.Wasm
	k.IterateAllDataRequestWasms(ctx, func(wasm types.Wasm) bool {
		wasms = append(wasms, wasm)
		return false
	})
	k.IterateAllOverlayWasms(ctx, func(wasm types.Wasm) bool {
		wasms = append(wasms, wasm)
		return false
	})
	return wasms
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
