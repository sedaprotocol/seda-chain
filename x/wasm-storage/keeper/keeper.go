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
func (k Keeper) SetDataRequestWasm(ctx sdk.Context, data, hash []byte) {
	ctx.KVStore(k.storeKey).Set(types.GetDataRequestWasmKey(hash), data)
}

// GetDataRequestWasm returns Data Request Wasm given its key.
func (k Keeper) GetDataRequestWasm(ctx sdk.Context, hash []byte) []byte {
	return ctx.KVStore(k.storeKey).Get(types.GetDataRequestWasmKey(hash))
}

// SetOverlayWasm stores Overlay Wasm using its hash as the key.
func (k Keeper) SetOverlayWasm(ctx sdk.Context, data, hash []byte) {
	ctx.KVStore(k.storeKey).Set(types.GetOverlayWasmKey(hash), data)
}

// GetOverlayWasm returns Overlay Wasm given its key.
func (k Keeper) GetOverlayWasm(ctx sdk.Context, hash []byte) []byte {
	return ctx.KVStore(k.storeKey).Get(types.GetOverlayWasmKey(hash))
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
