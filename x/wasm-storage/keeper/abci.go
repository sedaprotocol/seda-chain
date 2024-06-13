package keeper

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-wasm-vm/tallyvm"
)

type Request struct {
	DrBinaryID        string                 `json:"dr_binary_id"`
	DrInputs          string                 `json:"dr_inputs"`
	GasLimit          string                 `json:"gas_limit"`
	GasPrice          string                 `json:"gas_price"`
	Height            uint64                 `json:"height"`
	ID                string                 `json:"id"`
	Memo              string                 `json:"memo"`
	PaybackAddress    string                 `json:"payback_address"`
	ReplicationFactor int64                  `json:"replication_factor"`
	Reveals           map[string]interface{} `json:"reveals"`
	SedaPayload       string                 `json:"seda_payload"`
	TallyBinaryID     string                 `json:"tally_binary_id"`
	TallyInputs       string                 `json:"tally_inputs"`
	Version           string                 `json:"version"`
}

type TallyingList map[string]Request

type tallyArg struct {
	Reveals  map[string]any
	Outliers []bool
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
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, sdk.MustAccAddressFromBech32(contractAddr), []byte(`{"get_data_requests_by_status":{"status": "tallying"}}`))
	if err != nil {
		return err
	}
	var tallyingList TallyingList
	err = json.Unmarshal(queryRes, &tallyingList)
	if err != nil {
		return err
	}

	// 3. Loop through the list to apply filter and execute tally.
	// TODO: is it ok to use a map?
<<<<<<< HEAD
	for id := range tallyingList {
		// TODO: filtering

		tallyID, err := hex.DecodeString(tallyingList[id].TallyBinaryID)
		if err != nil {
			return fmt.Errorf("failed to decode tally ID to hex: %w", err)
		}
		tallyWasm, err := k.DataRequestWasm.Get(ctx, tallyID)
=======
	for id, req := range tallyingList {
		keys := make([]string, 0, len(req.Reveals))
		for k := range req.Reveals {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Sort reveals.
		reveals := make([]RevealBody, 0, len(req.Reveals))
		for _, k := range keys {
			reveals = append(reveals, req.Reveals[k])
		}
		outliers, consensus, err := Outliers(req.TallyInputs, reveals)
		if err != nil {
			return err
		}
		tallyWasm, err := k.DataRequestWasm.Get(ctx, req.TallyBinaryID)
>>>>>>> 325b97e (feat: none filter for DR wasm)
		if err != nil {
			return fmt.Errorf("failed to get tally wasm for DR ID %s: %w", id, err)
		}

		args, err := tallyVMArg(req.TallyInputs, req.Reveals, outliers)
		if err != nil {
			return err
		}

		result := tallyvm.ExecuteTallyVm(tallyWasm.Bytecode, args, map[string]string{
			"CONSENSUS": fmt.Sprintf("%v", consensus),
		})
		fmt.Println(result)
	}

	// 4. Post results.
	// msg := []byte("{\"data_requests\": {}}")
	// drContractAddr, err := k.wasmKeeper.Sudo(ctx, sdk.MustAccAddressFromBech32(proxyContractAddr), msg)
	// fmt.Println("dr contract addy: " + string(drContractAddr))

	return nil
}

func tallyVMArg(inputArgs []byte, reveals map[string]RevealBody, outliers []bool) ([]string, error) {
	arg := []string{string(inputArgs)}
	r, err := json.Marshal(reveals)
	if err != nil {
		return nil, err
	}
	arg = append(arg, string(r))
	o, err := json.Marshal(outliers)
	if err != nil {
		return nil, err
	}
	arg = append(arg, string(o))

	return arg, err
}
