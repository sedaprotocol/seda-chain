package keeper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) ExecuteTallyProgram(ctx sdk.Context, req types.Request, filterResult FilterResult, reveals []types.RevealBody, gasMeter *types.GasMeter) (types.VMResult, error) {
	tallyProgram, err := k.wasmStorageKeeper.GetOracleProgram(ctx, req.TallyProgramID)
	if err != nil {
		return types.VMResult{}, k.logErrAndRet(ctx, err, types.ErrFindingTallyProgram, req)
	}
	tallyInputs, err := base64.StdEncoding.DecodeString(req.TallyInputs)
	if err != nil {
		return types.VMResult{}, k.logErrAndRet(ctx, err, types.ErrDecodingTallyInputs, req)
	}

	// Convert base64-encoded payback address to hex encoding that
	// the tally VM expects.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.PaybackAddress)
	if err != nil {
		return types.VMResult{}, k.logErrAndRet(ctx, err, types.ErrDecodingPaybackAddress, req)
	}
	paybackAddrHex := hex.EncodeToString(decodedBytes)

	args, err := tallyVMArg(tallyInputs, reveals, filterResult.Outliers)
	if err != nil {
		return types.VMResult{}, k.logErrAndRet(ctx, err, types.ErrConstructingTallyVMArgs, req)
	}

	k.Logger(ctx).Info(
		"executing tally VM",
		"request_id", req.ID,
		"tally_program_id", req.TallyProgramID,
		"consensus", filterResult.Consensus,
		"arguments", args,
	)

	vmRes := tallyvm.ExecuteTallyVm(tallyProgram.Bytecode, args, map[string]string{
		"VM_MODE":               "tally",
		"CONSENSUS":             fmt.Sprintf("%v", filterResult.Consensus),
		"BLOCK_HEIGHT":          fmt.Sprintf("%d", ctx.BlockHeight()),
		"DR_ID":                 req.ID,
		"DR_REPLICATION_FACTOR": fmt.Sprintf("%v", req.ReplicationFactor),
		"EXEC_PROGRAM_ID":       req.ExecProgramID,
		"EXEC_INPUTS":           req.ExecInputs,
		"EXEC_GAS_LIMIT":        fmt.Sprintf("%v", req.ExecGasLimit),
		"TALLY_INPUTS":          req.TallyInputs,
		"TALLY_PROGRAM_ID":      req.TallyProgramID,
		"DR_TALLY_GAS_LIMIT":    fmt.Sprintf("%v", gasMeter.RemainingTallyGas()),
		"DR_GAS_PRICE":          req.GasPrice,
		"DR_MEMO":               req.Memo,
		"DR_PAYBACK_ADDRESS":    paybackAddrHex,
	})

	gasMeter.ConsumeTallyGas(vmRes.GasUsed)

	result := types.MapVMResult(vmRes)
	if result.ExitMessage != "" {
		k.Logger(ctx).Error("tally vm exit message", "request_id", req.ID, "exit_message", result.ExitMessage)
	}

	return result, nil
}

func tallyVMArg(inputArgs []byte, reveals []types.RevealBody, outliers []bool) ([]string, error) {
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
