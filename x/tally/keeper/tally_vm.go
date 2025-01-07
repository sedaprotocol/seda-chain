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

const (
	TallyExitCodeNotEnoughCommits   = 200 // tally VM not executed due to not enough commits
	TallyExitCodeNotEnoughReveals   = 201 // tally VM not executed due to not enough reveals
	TallyExitCodeInvalidFilterInput = 253 // tally VM not executed due to invalid filter input
	TallyExitCodeFilterError        = 254 // tally VM not executed due to filter error
	TallyExitCodeExecError          = 255 // error while executing tally VM
)

func (k Keeper) ExecuteTallyProgram(ctx sdk.Context, req types.Request, filterResult FilterResult, reveals []types.RevealBody) (tallyvm.VmResult, error) {
	tallyProgram, err := k.wasmStorageKeeper.GetOracleProgram(ctx, req.TallyProgramID)
	if err != nil {
		return tallyvm.VmResult{}, k.logErrAndRet(ctx, err, types.ErrFindingTallyProgram, req)
	}
	tallyInputs, err := base64.StdEncoding.DecodeString(req.TallyInputs)
	if err != nil {
		return tallyvm.VmResult{}, k.logErrAndRet(ctx, err, types.ErrDecodingTallyInputs, req)
	}

	// Convert base64-encoded payback address to hex encoding that
	// the tally VM expects.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.PaybackAddress)
	if err != nil {
		return tallyvm.VmResult{}, k.logErrAndRet(ctx, err, types.ErrDecodingPaybackAddress, req)
	}
	paybackAddrHex := hex.EncodeToString(decodedBytes)

	// Adjust gas limit based on the gas used by the filter.
	maxGasLimit, err := k.GetMaxTallyGasLimit(ctx)
	if err != nil {
		return tallyvm.VmResult{}, k.logErrAndRet(ctx, err, types.ErrGettingMaxTallyGasLimit, req)
	}
	var gasLimit uint64
	if min(req.TallyGasLimit, maxGasLimit) > filterResult.GasUsed {
		gasLimit = min(req.TallyGasLimit, maxGasLimit) - filterResult.GasUsed
	} else {
		gasLimit = 0
	}

	args, err := tallyVMArg(tallyInputs, reveals, filterResult.Outliers)
	if err != nil {
		return tallyvm.VmResult{}, k.logErrAndRet(ctx, err, types.ErrConstructingTallyVMArgs, req)
	}

	k.Logger(ctx).Info(
		"executing tally VM",
		"request_id", req.ID,
		"tally_program_id", req.TallyProgramID,
		"consensus", filterResult.Consensus,
		"arguments", args,
	)

	return tallyvm.ExecuteTallyVm(tallyProgram.Bytecode, args, map[string]string{
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
		"DR_TALLY_GAS_LIMIT":    fmt.Sprintf("%v", gasLimit),
		"DR_GAS_PRICE":          req.GasPrice,
		"DR_MEMO":               req.Memo,
		"DR_PAYBACK_ADDRESS":    paybackAddrHex,
	}), nil
}

func tallyVMArg(inputArgs []byte, reveals []types.RevealBody, outliers []bool) ([]string, error) {
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
