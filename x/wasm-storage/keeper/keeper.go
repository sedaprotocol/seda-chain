package keeper

import (
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	authority  string
	wasmKeeper wasmtypes.ContractOpsKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, authority string, wk wasmtypes.ContractOpsKeeper) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		authority:  authority,
		wasmKeeper: wk,
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
	var wasm types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDataRequestWasmKey(hash))
	k.cdc.MustUnmarshal(bz, &wasm)
	return &wasm
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
	var wasm types.Wasm
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetOverlayWasmKey(hash))
	k.cdc.MustUnmarshal(bz, &wasm)
	return &wasm
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
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixDataRequest)

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
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixOverlay)

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

// GetParams returns all the parameters for the module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(types.KeyParams))

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the parameters in the store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set([]byte(types.KeyParams), bz)
}

// SetMaxWasmSize updates the MaxWasmSize parameter.
func (k Keeper) SetMaxWasmSize(ctx sdk.Context, maxWasmSize uint64) error {
	if maxWasmSize == 0 {
		return errors.New("MaxWasmSize cannot be zero")
	}

	params := k.GetParams(ctx)
	params.MaxWasmSize = maxWasmSize
	k.SetParams(ctx, params)

	return nil
}
