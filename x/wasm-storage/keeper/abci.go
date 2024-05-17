package keeper

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vm "github.com/sedaprotocol/seda-wasm-vm/bind_go"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	file := "/Users/hykim/dev/seda-chain/x/wasm-storage/keeper/test_utils/debug.wasm"
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	result := vm.ExecuteTallyVm(data, []string{"1", "2"}, map[string]string{
		"PATH": os.Getenv("SHELL"),
	})
	fmt.Println(result)
	return nil
}
