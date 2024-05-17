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

type Keeper struct {
	authority  string
	wasmKeeper wasmtypes.ContractOpsKeeper

	// state management
	Schema                collections.Schema
	DataRequestWasm       collections.Map[[]byte, types.Wasm]
	OverlayWasm           collections.Map[[]byte, types.Wasm]
	WasmExpiration        collections.KeySet[collections.Pair[int64, []byte]]
	ProxyContractRegistry collections.Item[string]
	Params                collections.Item[types.Params]
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService, authority string, wk wasmtypes.ContractOpsKeeper) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	return &Keeper{
		authority:             authority,
		wasmKeeper:            wk,
		DataRequestWasm:       collections.NewMap(sb, types.DataRequestPrefix, "data-request-wasm", collections.BytesKey, codec.CollValue[types.Wasm](cdc)),
		OverlayWasm:           collections.NewMap(sb, types.OverlayPrefix, "overlay-wasm", collections.BytesKey, codec.CollValue[types.Wasm](cdc)),
		WasmExpiration:        collections.NewKeySet(sb, types.WasmExpPrefix, "wasm-exp", collections.PairKeyCodec(collections.Int64Key, collections.BytesKey)),
		ProxyContractRegistry: collections.NewItem(sb, types.ProxyContractRegistryPrefix, "proxy-contract-registry", collections.StringValue),
		Params:                collections.NewItem(sb, types.ParamsPrefix, "params", codec.CollValue[types.Params](cdc)),
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
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

// GetExpiredWasmKeys retrieves the keys of the data request wasms that
// will expire at the given block height.
func (k Keeper) GetExpiredWasmKeys(ctx sdk.Context, expirationHeight int64) ([][]byte, error) {
	ret := make([][]byte, 0)
	rng := collections.NewPrefixedPairRange[int64, []byte](expirationHeight)

	itr, err := k.WasmExpiration.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}

	keys, err := itr.Keys()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		ret = append(ret, k.K2())
	}
	return ret, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
