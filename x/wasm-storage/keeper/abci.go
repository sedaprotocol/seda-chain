package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	vm "github.com/sedaprotocol/seda-wasm-vm/bind_go"
)

type Request struct {
	Commits           map[string]interface{} `json:"commits"`
	DrBinaryID        []byte                 `json:"dr_binary_id"`
	DrInputs          []interface{}          `json:"dr_inputs"`
	GasLimit          string                 `json:"gas_limit"`
	GasPrice          string                 `json:"gas_price"`
	ID                []byte                 `json:"id"`
	Memo              []interface{}          `json:"memo"`
	PaybackAddress    []interface{}          `json:"payback_address"`
	ReplicationFactor int                    `json:"replication_factor"`
	Reveals           map[string]interface{} `json:"reveals"`
	SedaPayload       []interface{}          `json:"seda_payload"`
	TallyBinaryID     []byte                 `json:"tally_binary_id"`
	TallyInputs       []int                  `json:"tally_inputs"`
	Version           string                 `json:"version"`
}

type TallyingList map[string]struct {
	Request `json:"request"`
}

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
	// TODO: json marshal request struct?
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, sdk.MustAccAddressFromBech32(contractAddr), []byte(`{"get_data_requests_by_status":{"status": "tallying"}}`))
	if err != nil {
		return err
	}
	var tallyingList TallyingList
	err = json.Unmarshal(queryRes, &tallyingList)
	if err != nil {
		return err
	}

	// 3. Loop through the list to execute tally.
	// TODO: is it ok to use a map?
	for id := range tallyingList {
		// TODO: filtering

		tallyWasm, err := k.DataRequestWasm.Get(ctx, tallyingList[id].TallyBinaryID)
		if err != nil {
			return fmt.Errorf("failed to get tally wasm for DR ID %s: %w", id, err)
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
