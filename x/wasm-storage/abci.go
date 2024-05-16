package wasmstorage

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
)

// EndBlocker removes all the wasm that are expired at this block height.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	blockHeight := ctx.BlockHeight()
	keys, err := k.GetExpiredWasmKeys(ctx, blockHeight)
	if err != nil {
		return err
	}
	for _, wasmHash := range keys {
		if err := k.DataRequestWasm.Remove(ctx, wasmHash); err != nil {
			return err
		}
		if err := k.WasmExp.Remove(ctx, collections.Join(blockHeight, wasmHash)); err != nil {
			return err
		}
	}
	return nil
}
