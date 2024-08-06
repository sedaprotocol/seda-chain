package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in wasm-storage EndBlock", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in wasm-storage EndBlock", "err", err)
		}
		err = nil
	}()

	err = k.ProcessTallies(ctx)
	if err != nil {
		return
	}
	return
}

// ProcessTallies fetches from the core contract the list of requests
// to be tallied and then goes through it to filter and tally.
func (k Keeper) ProcessTallies(ctx sdk.Context) error {
	tallyWasm, err := os.ReadFile("./testwasm.wasm")
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(
		"About to execute tally wasmvm",
		len(tallyWasm),
	)

	vmRes := tallyvm.ExecuteTallyVm(tallyWasm, []string{"asdf"}, map[string]string{
		"VM_MODE":   "tally",
		"CONSENSUS": fmt.Sprintf("%v", true),
	})

	k.Logger(ctx).Info(
		"posting execution results to SEDA contract",
		"execution_result", vmRes,
	)

	k.Logger(ctx).Info(
		"tally flow completed",
	)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTallyCompletion,
			sdk.NewAttribute(types.AttributeTallyVMStdOut, strings.Join(vmRes.Stdout, "\n")),
			sdk.NewAttribute(types.AttributeTallyVMStdErr, strings.Join(vmRes.Stderr, "\n")),
		),
	)

	return nil
}

func tallyVMArg(inputArgs []byte, reveals []types.RevealBody, outliers []int) ([]string, error) {
	arg := []string{hex.EncodeToString(inputArgs)}

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
