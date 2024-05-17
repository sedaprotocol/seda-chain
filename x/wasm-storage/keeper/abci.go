package keeper

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	vm "github.com/sedaprotocol/seda-wasm-vm/bind_go"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	err := k.ProcessExpiredWasms(ctx)
	if err != nil {
		return err
	}

	err = k.ExecuteTally(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) ProcessExpiredWasms(ctx sdk.Context) error {
	blockHeight := ctx.BlockHeight()
	keys, err := k.GetExpiredWasmKeys(ctx, blockHeight)
	if err != nil {
		return err
	}
	for _, wasmHash := range keys {
		if err := k.DataRequestWasm.Remove(ctx, wasmHash); err != nil {
			return err
		}
		if err := k.WasmExpiration.Remove(ctx, collections.Join(blockHeight, wasmHash)); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) ExecuteTally(ctx sdk.Context) error {
	hash, err := hex.DecodeString("aad4d8a759c33a28bd6f6213c60e4e2f64d690ab559fc62d272a7d278170b802")
	if err != nil {
		return err
	}
	tallyWasm, err := k.DataRequestWasm.Get(ctx, hash)
	if err != nil {
		// TODO: reactivate error handling
		if errors.Is(err, collections.ErrNotFound) {
			return nil
		}
		return err
	}

	result := vm.ExecuteTallyVm(tallyWasm.Bytecode, []string{"1", "2"}, map[string]string{
		"PATH": os.Getenv("SHELL"),
	})
	fmt.Println(result)
	return nil
}
