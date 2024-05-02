package keeper

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var (
	// DataRequestPrefix defines prefix to store Data Request Wasm binaries.
	DataRequestPrefix = collections.NewPrefix(0)

	// OverlayPrefix defines prefix to store Overlay Wasm binaries.
	OverlayPrefix = collections.NewPrefix(1)

	ParamsPrefix = collections.NewPrefix(2)
)

func GetDataRequestWasmKeyPrefixFull(hash []byte) []byte {
	return append(DataRequestPrefix, hash...)
}

func GetOverlayWasmKeyPrefixFull(hash []byte) []byte {
	return append(OverlayPrefix, hash...)
}

type Keeper struct {
	authority  string
	wasmKeeper wasmtypes.ContractOpsKeeper

	// state management
	Schema                collections.Schema
	DataRequestWasm       collections.Map[[]byte, types.Wasm]
	OverlayWasm           collections.Map[[]byte, types.Wasm]
	ProxyContractRegistry collections.Item[[]byte]
	Params                collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string, wk wasmtypes.ContractOpsKeeper) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	return &Keeper{
		authority:       authority,
		wasmKeeper:      wk,
		DataRequestWasm: collections.NewMap(sb, DataRequestPrefix, "data-request-wasm", collections.BytesKey, codec.CollValue[types.Wasm](cdc)),
		OverlayWasm:     collections.NewMap(sb, OverlayPrefix, "overlay-wasm", collections.BytesKey, codec.CollValue[types.Wasm](cdc)),
		Params:          collections.NewItem(sb, ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetDataRequestWasm stores Data Request Wasm using its hash as the key.
func (k Keeper) SetDataRequestWasm(ctx sdk.Context, wasm *types.Wasm) error {
	dataRequestWasmKeyPrefixFull := GetDataRequestWasmKeyPrefixFull(wasm.Hash)

	return k.DataRequestWasm.Set(ctx, dataRequestWasmKeyPrefixFull, *wasm)
}

// GetDataRequestWasm returns Data Request Wasm given its key.
func (k Keeper) GetDataRequestWasm(ctx sdk.Context, hash []byte) (*types.Wasm, error) {
	dataRequestWasmKeyPrefixFull := GetDataRequestWasmKeyPrefixFull(hash)
	wasm, err := k.DataRequestWasm.Get(ctx, dataRequestWasmKeyPrefixFull)
	if err != nil {
		return nil, err
	}

	return &wasm, nil
}

// HasDataRequestWasm checks if a given Data Request Wasm exists.
func (k Keeper) HasDataRequestWasm(ctx sdk.Context, wasm *types.Wasm) (bool, error) {
	dataRequestWasmKeyPrefixFull := GetDataRequestWasmKeyPrefixFull(wasm.Hash)

	return k.DataRequestWasm.Has(ctx, dataRequestWasmKeyPrefixFull)
}

// SetOverlayWasm stores Overlay Wasm using its hash as the key.
func (k Keeper) SetOverlayWasm(ctx sdk.Context, wasm *types.Wasm) error {
	overlayWasmKeyPrefixFull := GetOverlayWasmKeyPrefixFull(wasm.Hash)
	return k.OverlayWasm.Set(ctx, overlayWasmKeyPrefixFull, *wasm)
}

// GetOverlayWasm returns Overlay Wasm given its key.
func (k Keeper) GetOverlayWasm(ctx sdk.Context, hash []byte) (*types.Wasm, error) {
	overlayWasmKeyPrefixFull := GetOverlayWasmKeyPrefixFull(hash)
	wasm, err := k.OverlayWasm.Get(ctx, overlayWasmKeyPrefixFull)
	if err != nil {
		return nil, err
	}

	return &wasm, nil
}

// HasOverlayWasm checks if a given Overlay Wasm exists.
func (k Keeper) HasOverlayWasm(ctx sdk.Context, wasm *types.Wasm) (bool, error) {
	overlayWasmKeyPrefixFull := GetOverlayWasmKeyPrefixFull(wasm.Hash)

	return k.OverlayWasm.Has(ctx, overlayWasmKeyPrefixFull)
}

// IterateAllDataRequestWasms iterates over the all the stored Data Request
// Wasms and performs a given callback function.
func (k Keeper) IterateAllDataRequestWasms(ctx sdk.Context, callback func(wasm types.Wasm) (stop bool)) error {
	iter, err := k.DataRequestWasm.Iterate(ctx, nil)
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

// IterateAllOverlayWasms iterates over the all the stored Overlay Wasms
// and performs a given callback function.
func (k Keeper) IterateAllOverlayWasms(ctx sdk.Context, callback func(wasm types.Wasm) (stop bool)) error {
	iter, err := k.OverlayWasm.Iterate(ctx, nil)
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

// ListDataRequestWasms returns hashes and types of all Data Request Wasms
// in the store.
func (k Keeper) ListDataRequestWasms(ctx sdk.Context) []string {
	var hashTypePairs []string
	err := k.IterateAllDataRequestWasms(ctx, func(w types.Wasm) bool {
		hashTypePairs = append(hashTypePairs, hex.EncodeToString(w.Hash)+","+w.WasmType.String())
		return false
	})
	if err != nil {
		return nil
	}
	return hashTypePairs
}

// ListOverlayWasms returns hashes and types of all Overlay Wasms in the store.
func (k Keeper) ListOverlayWasms(ctx sdk.Context) []string {
	var hashTypePairs []string
	err := k.IterateAllOverlayWasms(ctx, func(w types.Wasm) bool {
		hashTypePairs = append(hashTypePairs, hex.EncodeToString(w.Hash)+","+w.WasmType.String())
		return false
	})
	if err != nil {
		return nil
	}
	return hashTypePairs
}

func (k Keeper) GetAllWasms(ctx sdk.Context) []types.Wasm {
	var wasms []types.Wasm
	err := k.IterateAllDataRequestWasms(ctx, func(wasm types.Wasm) bool {
		wasms = append(wasms, wasm)
		return false
	})
	if err != nil {
		return nil
	}
	err = k.IterateAllOverlayWasms(ctx, func(wasm types.Wasm) bool {
		wasms = append(wasms, wasm)
		return false
	})
	if err != nil {
		return nil
	}
	return wasms
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
