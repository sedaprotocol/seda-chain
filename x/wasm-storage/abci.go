package wasmstorage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
)

// EndBlocker removes all the wasm that are expired at this block height.
// if we are to add more functionality to EndBlocker, we have to factor out
// the wasm pruning into a separate function. The way code is structured for
// EndBlocker is sufficient for now.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	keys, err := k.WasmKeyByExpBlock(ctx, ctx.BlockHeight())
	if err != nil {
		return err
	}
	for _, wasmHash := range keys {
		if err := k.DataRequestWasm.Remove(ctx, wasmHash); err != nil {
			return err
		}
	}
	return nil
}
