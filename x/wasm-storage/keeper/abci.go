package keeper

import (
	"encoding/base64"
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
	DrBinaryID        string                `json:"dr_binary_id"`
	DrInputs          string                `json:"dr_inputs"`
	GasLimit          string                `json:"gas_limit"`
	GasPrice          string                `json:"gas_price"`
	Height            uint64                `json:"height"`
	ID                string                `json:"id"`
	Memo              string                `json:"memo"`
	PaybackAddress    string                `json:"payback_address"`
	ReplicationFactor int64                 `json:"replication_factor"`
	Reveals           map[string]RevealBody `json:"reveals"`
	SedaPayload       string                `json:"seda_payload"`
	TallyBinaryID     string                `json:"tally_binary_id"`
	TallyInputs       string                `json:"tally_inputs"`
	Version           string                `json:"version"`
}

type RevealBody struct {
	ExitCode uint8  `json:"exit_code"`
	Reveal   string `json:"reveal"` // base64-encoded string
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
	var tallyList []Request
	err = json.Unmarshal(queryRes, &tallyList)
	if err != nil {
		return err
	}

	// 3. Loop through the list to apply filter and execute tally.
	for id, req := range tallyList {
		tallyInputs, err := base64.StdEncoding.DecodeString(req.TallyInputs)
		if err != nil {
			return fmt.Errorf("failed to decode tally input: %w", err)
		}

		// Sort reveals.
		keys := make([]string, len(req.Reveals))
		i := 0
		for k := range req.Reveals {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		reveals := make([]RevealBody, len(req.Reveals))
		for i, k := range keys {
			reveals[i] = req.Reveals[k]
		}

		outliers, consensus, err := ApplyFilter(tallyInputs, reveals)
		if err != nil {
			return err
		}

		tallyID, err := hex.DecodeString(req.TallyBinaryID)
		if err != nil {
			return fmt.Errorf("failed to decode tally ID to hex: %w", err)
		}
		tallyWasm, err := k.DataRequestWasm.Get(ctx, tallyID)
		if err != nil {
			return fmt.Errorf("failed to get tally wasm for DR ID %d: %w", id, err)
		}

		args, err := tallyVMArg(tallyInputs, req.Reveals, outliers)
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
