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
	// 1. Get contract address.
	// TODO: query contract to retrieve list of data requests
	// ready to be tallied and obtain their associated tally
	// wasm IDs.
	//
	proxyContractAddr, err := k.ProxyContractRegistry.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			k.Logger(ctx).Debug("proxy contract address not registered")
			return nil
		}
		return err
	}

	// msg := []byte("{\"data_requests\": {}}")
	// drContractAddr, err := k.wasmKeeper.Sudo(ctx, sdk.MustAccAddressFromBech32(proxyContractAddr), msg)
	// fmt.Println("dr contract addy: " + string(drContractAddr))

	fmt.Println(proxyContractAddr)
	// drContractAddr := proxyContractAddr

	// 2. Fetch tally-ready DRs.
	// bankQuery := wasmvmtypes.QueryRequest{
	// 	Bank: &wasmvmtypes.BankQuery{
	// 		AllBalances: &wasmvmtypes.AllBalancesQuery{
	// 			Address: creator.String(),
	// 		},
	// 	},
	// }
	// simpleQueryBz, err := json.Marshal(testdata.ReflectQueryMsg{
	// 	Chain: &testdata.ChainQuery{Request: &bankQuery},
	// })
	// if err != nil {
	// 	return err
	// }
	// list, err := k.wasmViewKeeper.QuerySmart(ctx, sdk.AccAddress(drContractAddr), simpleQueryBz)
	// // list, err := k.wasmViewKeeper.QuerySmart(ctx, sdk.AccAddress(drContractAddr), []byte{`{"verifier":{}}`})
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(list)

	// 3. Loop through the list to execute tally.
	// for _, item := range list {

	// }

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

	// 4. Post result.

	return nil
}
