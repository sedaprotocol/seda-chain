package keeper

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
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
	contractAddr, err := k.ProxyContractRegistry.Get(ctx)
	if contractAddr == "" || errors.Is(err, collections.ErrNotFound) {
		k.Logger(ctx).Debug("proxy contract address not registered")
		return nil
	}
	if err != nil {
		return err
	}

	// 2. Fetch tally-ready DRs.
	// TODO: json marshal golang struct?
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, sdk.MustAccAddressFromBech32(contractAddr), []byte(`{"get_data_requests_by_status":{"status": "tallying"}}`))
	if err != nil {
		return err
	}
	fmt.Println(string(queryRes))

	type DataRequest struct {
		Id                string
		Version           string
		DrBinaryId        string
		DrInputs          []byte
		TallyBinaryId     string
		TallyInputs       []byte
		ReplicationFactor uint16
		GasPrice          *big.Int
		GasLimit          *big.Int
		Memo              string
		PaybackAddress    []byte
		SedaPayload       []byte
		Commits           map[string]string
		// Reveals           map[string]RevealBody
	}

	var hmap map[string]DataRequest
	err = json.Unmarshal(queryRes, &hmap)
	if err != nil {
		return err
	}

	// 3. Loop through the list to execute tally.
	// #[returns(HashMap<String, DR>)]
	// TODO: is it ok to use a map?

	for key := range hmap {
		hash, err := hex.DecodeString(hmap[key].TallyBinaryId)
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
	}

	// 4. Post result.
	// msg := []byte("{\"data_requests\": {}}")
	// drContractAddr, err := k.wasmKeeper.Sudo(ctx, sdk.MustAccAddressFromBech32(proxyContractAddr), msg)
	// fmt.Println("dr contract addy: " + string(drContractAddr))

	return nil
}
