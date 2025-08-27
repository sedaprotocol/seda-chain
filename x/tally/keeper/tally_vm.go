package keeper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v3"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

type TallyParallelExecItem struct {
	Request   types.Request
	GasMeter  *types.GasMeter
	Reveals   []types.Reveal
	Outliers  []bool
	Consensus bool
	// Index is the corresponding index in tallyResults and dataResults arrays.
	Index int
	// If TallyExecErr is not nil, the item was not executed due to this error.
	TallyExecErr error
}

func NewTallyParallelExecItem(index int, req types.Request, gasMeter *types.GasMeter, reveals []types.Reveal, outliers []bool, consensus bool) TallyParallelExecItem {
	return TallyParallelExecItem{
		Index:     index,
		Request:   req,
		GasMeter:  gasMeter,
		Reveals:   reveals,
		Outliers:  outliers,
		Consensus: consensus,
	}
}

// BatchExecuteTallyProgramsParallel executes ExecuteTallyProgramsParallel in
// batches. The batch size is currently set to 25.
func (k Keeper) BatchExecuteTallyProgramsParallel(ctx sdk.Context, tallyExecItems []TallyParallelExecItem) []tallyvm.VmResult {
	batchSize := 25
	var vmResults []tallyvm.VmResult
	for i := 0; i*batchSize < len(tallyExecItems); i++ {
		end := min((i+1)*batchSize, len(tallyExecItems))
		vmRes := k.ExecuteTallyProgramsParallel(ctx, tallyExecItems[i*batchSize:end])
		vmResults = append(vmResults, vmRes...)
	}

	return vmResults
}

// ExecuteTallyProgramParallel executes tally programs in parallel given a slice
// of TallyParallelExecItems that contain execution information.
// If an item is not executed due to an error, the error is recorded in the item.
// This method returns a slice of VM execution results of the items that are
// executed in order.
func (k Keeper) ExecuteTallyProgramsParallel(ctx sdk.Context, items []TallyParallelExecItem) []tallyvm.VmResult {
	programs := make([][]byte, 0, len(items))
	args := make([][]string, 0, len(items))
	envs := make([]map[string]string, 0, len(items))

	for i := range items {
		program, err := k.wasmStorageKeeper.GetOracleProgram(ctx, items[i].Request.TallyProgramID)
		if err != nil {
			items[i].TallyExecErr = err
			k.Logger(ctx).Error(err.Error(), "request_id", items[i].Request.ID, "error", types.ErrFindingTallyProgram)
			continue
		}

		input, err := base64.StdEncoding.DecodeString(items[i].Request.TallyInputs)
		if err != nil {
			items[i].TallyExecErr = err
			k.Logger(ctx).Error(err.Error(), "request_id", items[i].Request.ID, "error", types.ErrDecodingTallyInputs)
			continue
		}

		// Convert base64 payback address to hex that tally VM expects.
		paybackAddrBytes, err := base64.StdEncoding.DecodeString(items[i].Request.PaybackAddress)
		if err != nil {
			items[i].TallyExecErr = err
			k.Logger(ctx).Error(err.Error(), "request_id", items[i].Request.ID, "error", types.ErrDecodingPaybackAddress)
			continue
		}

		arg, err := tallyVMArg(input, items[i].Reveals, items[i].Outliers)
		if err != nil {
			items[i].TallyExecErr = err
			k.Logger(ctx).Error(err.Error(), "request_id", items[i].Request.ID, "error", types.ErrConstructingTallyVMArgs)
			continue
		}

		programs = append(programs, program.Bytecode)
		args = append(args, arg)
		envs = append(envs, map[string]string{
			"VM_MODE":               "tally",
			"CONSENSUS":             fmt.Sprintf("%v", items[i].Consensus),
			"BLOCK_HEIGHT":          fmt.Sprintf("%d", ctx.BlockHeight()),
			"DR_ID":                 items[i].Request.ID,
			"DR_REPLICATION_FACTOR": fmt.Sprintf("%v", items[i].Request.ReplicationFactor),
			"EXEC_PROGRAM_ID":       items[i].Request.ExecProgramID,
			"EXEC_INPUTS":           items[i].Request.ExecInputs,
			"EXEC_GAS_LIMIT":        fmt.Sprintf("%v", items[i].Request.ExecGasLimit),
			"TALLY_INPUTS":          items[i].Request.TallyInputs,
			"TALLY_PROGRAM_ID":      items[i].Request.TallyProgramID,
			"DR_TALLY_GAS_LIMIT":    fmt.Sprintf("%v", items[i].GasMeter.RemainingTallyGas()),
			"DR_GAS_PRICE":          items[i].Request.PostedGasPrice,
			"DR_MEMO":               items[i].Request.Memo,
			"DR_PAYBACK_ADDRESS":    hex.EncodeToString(paybackAddrBytes),
		})

		k.Logger(ctx).Debug(
			"executing tally VM",
			"request_id", items[i].Request.ID,
			"tally_program_id", items[i].Request.TallyProgramID,
		)
	}
	if len(programs) == 0 {
		return []tallyvm.VmResult{}
	}

	return tallyvm.ExecuteMultipleFromCParallel(programs, args, envs)
}

func tallyVMArg(inputArgs []byte, reveals []types.Reveal, outliers []bool) ([]string, error) {
	arg := []string{hex.EncodeToString(inputArgs)}

	r, err := json.Marshal(reveals)
	if err != nil {
		return nil, err
	}
	arg = append(arg, string(r))

	outliersArg := make([]int, len(outliers))
	for i, outlier := range outliers {
		if outlier {
			outliersArg[i] = 1
		} else {
			outliersArg[i] = 0
		}
	}

	o, err := json.Marshal(outliersArg)
	if err != nil {
		return nil, err
	}
	arg = append(arg, string(o))

	return arg, nil
}
