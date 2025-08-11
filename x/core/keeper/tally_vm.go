package keeper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (k Keeper) ExecuteTallyProgram(ctx sdk.Context, dr types.DataRequest, filterResult FilterResult, reveals []types.Reveal, gasMeter *types.GasMeter) (types.VMResult, error) {
	tallyProgram, err := k.wasmStorageKeeper.GetOracleProgram(ctx, dr.TallyProgramId)
	if err != nil {
		return types.VMResult{}, k.logErrAndRet(ctx, err, types.ErrFindingTallyProgram, dr.Id)
	}

	args, err := tallyVMArg(dr.TallyInputs, reveals, filterResult.Outliers)
	if err != nil {
		return types.VMResult{}, k.logErrAndRet(ctx, err, types.ErrConstructingTallyVMArgs, dr.Id)
	}

	k.Logger(ctx).Info(
		"executing tally VM",
		"request_id", dr.Id,
		"tally_program_id", dr.TallyProgramId,
		"consensus", filterResult.Consensus,
		"arguments", args,
	)

	vmRes := tallyvm.ExecuteTallyVm(tallyProgram.Bytecode, args, map[string]string{
		"VM_MODE":               "tally",
		"CONSENSUS":             fmt.Sprintf("%v", filterResult.Consensus),
		"BLOCK_HEIGHT":          fmt.Sprintf("%d", ctx.BlockHeight()),
		"DR_ID":                 dr.Id,
		"DR_REPLICATION_FACTOR": fmt.Sprintf("%d", dr.ReplicationFactor),
		"EXEC_PROGRAM_ID":       dr.ExecProgramId,
		"EXEC_INPUTS":           base64.StdEncoding.EncodeToString(dr.ExecInputs), // vm expects base64-encoded string
		"EXEC_GAS_LIMIT":        fmt.Sprintf("%d", dr.ExecGasLimit),
		"TALLY_INPUTS":          base64.StdEncoding.EncodeToString(dr.TallyInputs), // vm expects base64-encoded string
		"TALLY_PROGRAM_ID":      dr.TallyProgramId,
		"DR_TALLY_GAS_LIMIT":    fmt.Sprintf("%d", gasMeter.RemainingTallyGas()),
		"DR_GAS_PRICE":          dr.PostedGasPrice.String(),
		"DR_MEMO":               base64.StdEncoding.EncodeToString(dr.Memo), // vm expects base64-encoded string
		"DR_PAYBACK_ADDRESS":    hex.EncodeToString(dr.PaybackAddress),      // vm expects hex string
	})

	gasMeter.ConsumeTallyGas(vmRes.GasUsed)

	result := types.MapVMResult(vmRes)
	if result.ExitCode != 0 {
		k.Logger(ctx).Error("tally vm exit message", "request_id", dr.Id, "exit_message", result.ExitMessage)
	}

	return result, nil
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
