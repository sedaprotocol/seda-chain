package keeper

import (
	"errors"
	"fmt"
	"os"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vm "github.com/sedaprotocol/seda-wasm-vm/bind_go"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	fmt.Println("herererer")
	tallyWasm, err := k.DataRequestWasm.Get(ctx, append(DataRequestPrefix, []byte("aad4d8a759c33a28bd6f6213c60e4e2f64d690ab559fc62d272a7d278170b802")...))
	if err != nil {
		fmt.Println(err)

		if errors.Is(err, collections.ErrNotFound) {
			return nil
		}
		panic(err)
	}

	fmt.Println("les execute")

	// file := "/Users/hykim/dev/seda-chain/x/wasm-storage/keeper/test_utils/debug.wasm"
	// data, err := os.ReadFile(file)
	// if err != nil {
	// 	return err
	// }

	result := vm.ExecuteTallyVm(tallyWasm.Bytecode, []string{"1", "2"}, map[string]string{
		"PATH": os.Getenv("SHELL"),
	})
	fmt.Println(result)
	return nil
}
