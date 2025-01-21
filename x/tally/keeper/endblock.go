package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in tally end block", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in tally end block", "err", err)
		}
		err = nil
	}()

	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return
	}
	if coreContract == nil {
		k.Logger(ctx).Info("skipping tally end block - core contract has not been registered")
		return
	}

	postRes, err := k.wasmKeeper.Sudo(ctx, coreContract, []byte(`{"expire_data_requests":{}}`))
	if err != nil {
		return
	}
	k.Logger(ctx).Debug("sudo expire_data_requests", "res", postRes)

	err = k.ProcessTallies(ctx, coreContract)
	if err != nil {
		return
	}
	return
}

// ProcessTallies fetches from the core contract the list of requests
// to be tallied and then goes through it to filter and tally.
func (k Keeper) ProcessTallies(ctx sdk.Context, coreContract sdk.AccAddress) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	// Fetch tally-ready data requests.
	// TODO: Deal with offset and limits. (#313)
	queryRes, err := k.wasmViewKeeper.QuerySmart(ctx, coreContract, []byte(`{"get_data_requests_by_status":{"status": "tallying", "offset": 0, "limit": 100}}`))
	if err != nil {
		return err
	}
	if string(queryRes) == "[]" {
		return nil
	}
	k.Logger(ctx).Info("non-empty tally list - starting tally process")

	var tallyList []types.Request
	err = json.Unmarshal(queryRes, &tallyList)
	if err != nil {
		return err
	}

	// Loop through the list to apply filter, execute tally, and post
	// execution result.
	processedReqs := make(map[string][]types.Distribution)
	tallyResults := make([]TallyResult, len(tallyList))
	dataResults := make([]batchingtypes.DataResult, len(tallyList))
	for i, req := range tallyList {
		dataResults[i] = batchingtypes.DataResult{
			DrId:          req.ID,
			DrBlockHeight: req.Height,
			Version:       req.Version,
			//nolint:gosec // G115: We shouldn't get negative block heights anyway.
			BlockHeight: uint64(ctx.BlockHeight()),
			//nolint:gosec // G115: We shouldn't get negative timestamps anyway.
			BlockTimestamp: uint64(ctx.BlockTime().Unix()),
			PaybackAddress: req.PaybackAddress,
			SedaPayload:    req.SedaPayload,
		}

		gasPrice, ok := math.NewIntFromString(req.GasPrice)
		if !ok {
			return fmt.Errorf("invalid gas price: %s", req.GasPrice)
		}

		gasMeter := types.NewGasMeter(req.TallyGasLimit, req.ExecGasLimit, params.MaxTallyGasLimit, gasPrice, params.GasCostBase)
		if len(req.Commits) < int(req.ReplicationFactor) {
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", req.ReplicationFactor, len(req.Commits)))
			dataResults[i].ExitCode = TallyExitCodeNotEnoughCommits
			k.Logger(ctx).Info("data request's number of commits did not meet replication factor", "request_id", req.ID)

			MeterExecutorGasFallback(req, params.GasCostFallback, gasMeter, false)
		} else {
			_, tallyResults[i] = k.FilterAndTally(ctx, req, params, gasMeter)
			dataResults[i].Result = tallyResults[i].Result
			//nolint:gosec // G115: We shouldn't get negative exit code anyway.
			dataResults[i].ExitCode = uint32(tallyResults[i].ExitInfo.ExitCode)
			dataResults[i].Consensus = tallyResults[i].Consensus

			k.Logger(ctx).Info("completed tally", "request_id", req.ID)
			k.Logger(ctx).Debug("tally result", "request_id", req.ID, "tally_result", tallyResults[i])
		}

		processedReqs[req.ID] = k.DistributionsFromGasMeter(ctx, req.ID, req.Height, gasMeter, params.BurnRatio)

		dataResults[i].GasUsed = gasMeter.TotalGasUsed()
		dataResults[i].Id, err = dataResults[i].TryHash()
		if err != nil {
			return err
		}
	}

	// Notify the Core Contract of tally completion.
	msg, err := types.MarshalSudoRemoveDataRequests(processedReqs)
	if err != nil {
		return err
	}
	_, err = k.wasmKeeper.Sudo(ctx, coreContract, msg)
	if err != nil {
		return err
	}

	// Store the data results for batching.
	for i := range dataResults {
		err := k.batchingKeeper.SetDataResultForBatching(ctx, dataResults[i])
		if err != nil {
			return err
		}

		k.Logger(ctx).Info("tally flow completed", "request_id", dataResults[i].DrId)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTallyCompletion,
				sdk.NewAttribute(types.AttributeDataResultID, dataResults[i].Id),
				sdk.NewAttribute(types.AttributeDataRequestID, dataResults[i].DrId),
				sdk.NewAttribute(types.AttributeTypeConsensus, strconv.FormatBool(dataResults[i].Consensus)),
				sdk.NewAttribute(types.AttributeTallyVMStdOut, strings.Join(tallyResults[i].StdOut, "\n")),
				sdk.NewAttribute(types.AttributeTallyVMStdErr, strings.Join(tallyResults[i].StdErr, "\n")),
				sdk.NewAttribute(types.AttributeExecGasUsed, fmt.Sprintf("%v", tallyResults[i].ExecGasUsed)),
				sdk.NewAttribute(types.AttributeTallyGasUsed, fmt.Sprintf("%v", tallyResults[i].TallyGasUsed)),
				sdk.NewAttribute(types.AttributeTallyExitCode, fmt.Sprintf("%02x", dataResults[i].ExitCode)),
				sdk.NewAttribute(types.AttributeProxyPubKeys, strings.Join(tallyResults[i].ProxyPubKeys, "\n")),
			),
		)
	}

	return nil
}

type TallyResult struct {
	Consensus    bool
	StdOut       []string
	StdErr       []string
	Result       []byte
	ExitInfo     tallyvm.ExitInfo
	ExecGasUsed  uint64
	TallyGasUsed uint64
	ProxyPubKeys []string // data proxy pubkeys in basic consensus
}

// FilterAndTally builds and applies filter, executes tally program, and
// calculates canonical gas consumption.
func (k Keeper) FilterAndTally(ctx sdk.Context, req types.Request, params types.Params, gasMeter *types.GasMeter) (FilterResult, TallyResult) {
	// Sort the reveals by their keys (executors).
	keys := make([]string, len(req.Reveals))
	i := 0
	for k := range req.Reveals {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	reveals := make([]types.RevealBody, len(req.Reveals))
	for i, k := range keys {
		reveals[i] = req.Reveals[k]
		sort.Strings(reveals[i].ProxyPubKeys)
	}

	// Phase 1: Filtering
	filterResult, filterErr := ExecuteFilter(reveals, req.ConsensusFilter, req.ReplicationFactor, params, gasMeter)
	tallyResult := TallyResult{
		Consensus:    filterResult.Consensus,
		ProxyPubKeys: filterResult.ProxyPubKeys,
	}

	// Phase 2: Tally Program Execution
	var vmRes tallyvm.VmResult
	var tallyErr error
	if filterErr == nil {
		vmRes, tallyErr = k.ExecuteTallyProgram(ctx, req, filterResult, reveals, gasMeter)
		if tallyErr != nil {
			tallyResult.Result = []byte(tallyErr.Error())
			tallyResult.ExitInfo.ExitCode = TallyExitCodeExecError
		} else {
			tallyResult.Result = vmRes.Result
			tallyResult.ExitInfo = vmRes.ExitInfo
			tallyResult.StdOut = vmRes.Stdout
			tallyResult.StdErr = vmRes.Stderr
		}
	} else {
		tallyResult.Result = []byte(filterErr.Error())
		if errors.Is(filterErr, types.ErrInvalidFilterInput) {
			tallyResult.ExitInfo.ExitCode = TallyExitCodeInvalidFilterInput
		} else {
			tallyResult.ExitInfo.ExitCode = TallyExitCodeFilterError
		}
	}

	// Phase 3: Calculate data proxy and executor gas consumption.
	// Calculate data proxy gas consumption if basic consensus was reached.
	if filterErr == nil || !errors.Is(filterErr, types.ErrNoBasicConsensus) {
		k.MeterProxyGas(ctx, tallyResult.ProxyPubKeys, req.ReplicationFactor, gasMeter)
	}

	// Calculate executor gas consumption.
	switch {
	case errors.Is(filterErr, types.ErrNoBasicConsensus):
		MeterExecutorGasFallback(req, params.GasCostFallback, gasMeter, true)
	case errors.Is(filterErr, types.ErrInvalidFilterInput) || errors.Is(filterErr, types.ErrNoConsensus) || tallyErr != nil:
		gasMeter.SetReducedPayoutMode()
		fallthrough
	default: // filterErr == ErrConsensusInError || filterErr == nil
		proxyGasUsedPerExec := gasMeter.ExecutionGasUsed() / uint64(req.ReplicationFactor)
		gasReports := make([]uint64, len(reveals))
		for i, reveal := range reveals {
			if proxyGasUsedPerExec > reveal.GasUsed {
				gasReports[i] = 0
			} else {
				gasReports[i] = reveal.GasUsed - proxyGasUsedPerExec
			}
		}
		if areGasReportsUniform(gasReports) {
			MeterExecutorGasUniform(keys, gasReports[0], req.ReplicationFactor, gasMeter)
		} else {
			MeterExecutorGasDivergent(keys, gasReports, req.ReplicationFactor, gasMeter)
		}
	}

	tallyResult.TallyGasUsed = gasMeter.TallyGasUsed()
	tallyResult.ExecGasUsed = gasMeter.ExecutionGasUsed()
	return filterResult, tallyResult
}

// logErrAndRet logs the base error along with the request ID for
// debugging and returns the registered error.
func (k Keeper) logErrAndRet(ctx sdk.Context, baseErr, registeredErr error, req types.Request) error {
	k.Logger(ctx).Debug(baseErr.Error(), "request_id", req.ID, "error", registeredErr)
	return registeredErr
}
