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
	// Get contract address.
	contractAddrBech32, err := k.ProxyContractRegistry.Get(ctx)
	if contractAddrBech32 == "" || errors.Is(err, collections.ErrNotFound) {
		k.Logger(ctx).Debug("proxy contract address not registered")
		return nil
	}
	if err != nil {
		return err
	}
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	if err != nil {
		return err
	}

	// Fetch tally-ready data requests.
	// TODO: Deal with offset and limits.
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get_data_requests_by_status":{"status": "tallying", "offset": 0, "limit": 100}}`))
	if err != nil {
		return err
	}
	if string(queryRes) == "[]" {
		return nil
	}

	k.Logger(ctx).Info("non-empty tally list - starting tally process")

	// Loop through the list to apply filter, execute tally, and post
	// execution result.
	var tallyList []Request
	err = json.Unmarshal(queryRes, &tallyList)
	if err != nil {
		return err
	}
	for id, req := range tallyList {
		filter, err := base64.StdEncoding.DecodeString(req.ConsensusFilter)
		if err != nil {
			return fmt.Errorf("failed to decode consensus filter: %w", err)
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

		outliers, consensus, err := ApplyFilter(filter, reveals)
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
		tallyInputs, err := base64.StdEncoding.DecodeString(req.TallyInputs)
		if err != nil {
			return fmt.Errorf("failed to decode tally inputs: %w", err)
		}

		args, err := tallyVMArg(tallyInputs, reveals, outliers)
		if err != nil {
			return err
		}

		k.Logger(ctx).Info(
			"executing tally VM",
			"request_id", req.ID,
			"tally_wasm_hash", req.TallyBinaryID,
			"consensus", consensus,
			"arguments", args,
		)
		vmRes := tallyvm.ExecuteTallyVm(tallyWasm.Bytecode, args, map[string]string{
			"VM_MODE":   "tally",
			"CONSENSUS": fmt.Sprintf("%v", consensus),
		})

		// Post results to the SEDA contract.
		sudoMsg := Sudo{
			ID: req.ID,
			Result: DataResult{
				Version:        req.Version,
				ID:             req.ID,
				BlockHeight:    uint64(ctx.BlockHeight()),
				ExitCode:       byte(vmRes.ExitInfo.ExitCode),
				GasUsed:        "0", // TODO
				Result:         vmRes.Result,
				PaybackAddress: req.PaybackAddress,
				SedaPayload:    req.SedaPayload,
				Consensus:      consensus,
			},
			ExitCode: byte(vmRes.ExitInfo.ExitCode),
		}
		msg, err := json.Marshal(struct {
			PostDataResult Sudo `json:"post_data_result"`
		}{
			PostDataResult: sudoMsg,
		})
		if err != nil {
			return err
		}

		k.Logger(ctx).Info(
			"posting execution results to SEDA contract",
			"request_id", req.ID,
			"execution_result", vmRes,
			"sudo_message", sudoMsg,
		)
		postRes, err := k.wasmKeeper.Sudo(ctx, contractAddr, msg)
		if err != nil {
			return err
		}

		k.Logger(ctx).Info(
			"tally flow completed",
			"request_id", req.ID,
			"post_result", postRes,
		)
	}

	return nil
}

func tallyVMArg(inputArgs []byte, reveals []RevealBody, outliers []int) ([]string, error) {
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
