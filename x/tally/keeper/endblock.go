package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get core contract address", "err", err)
		return nil
	}
	if coreContract == nil {
		k.Logger(ctx).Info("skipping tally end block - core contract has not been registered")
		return nil
	}

	postRes, err := k.wasmKeeper.Sudo(ctx, coreContract, []byte(`{"expire_data_requests":{}}`))
	if err != nil {
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to expire data requests", "err", err)
		return nil
	}
	k.Logger(ctx).Debug("sudo expire_data_requests", "res", postRes)

	return k.ProcessTallies(ctx, coreContract)
}

// ProcessTallies fetches from the core contract the list of requests
// to be tallied and then goes through it to filter and tally.
func (k Keeper) ProcessTallies(ctx sdk.Context, coreContract sdk.AccAddress) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tally params", "err", err)
		return nil
	}
	tallyvm.TallyMaxBytes = uint(params.MaxResultSize)

	// Fetch tally-ready data requests.
	queryRes, err := k.wasmViewKeeper.QuerySmart(
		ctx, coreContract,
		[]byte(fmt.Sprintf(`{"get_data_requests_by_status":{"status": "tallying", "offset": 0, "limit": %d}}`, params.MaxTalliesPerBlock)),
	)
	if err != nil {
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tally-ready data requests", "err", err)
		return nil
	}

	var contractQueryResponse types.ContractListResponse
	err = json.Unmarshal(queryRes, &contractQueryResponse)
	if err != nil {
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to unmarshal data requests contract response", "err", err)
		return nil
	}

	tallyList := contractQueryResponse.DataRequests
	if len(tallyList) == 0 {
		k.Logger(ctx).Debug("no tally-ready data requests - skipping tally process")
		return nil
	}
	k.Logger(ctx).Info("non-empty tally list - starting tally process")

	// Loop through the list to apply filter, execute tally, and post
	// execution result.
	processedReqs := make(map[string][]types.Distribution)
	tallyResults := make([]TallyResult, len(tallyList))
	dataResults := make([]batchingtypes.DataResult, len(tallyList))

	for i, req := range tallyList {
		dataResults[i], err = req.ToResult(ctx)
		if err != nil {
			types.MarkResultAsFallback(&dataResults[i], err)
			continue
		}

		if contractQueryResponse.IsPaused {
			types.MarkResultAsPaused(&dataResults[i])
			continue
		}

		gasPrice, ok := math.NewIntFromString(req.GasPrice)
		if !ok {
			types.MarkResultAsFallback(&dataResults[i], fmt.Errorf("invalid gas price: %s", req.GasPrice))
			continue
		}

		gasMeter := types.NewGasMeter(req.TallyGasLimit, req.ExecGasLimit, params.MaxTallyGasLimit, gasPrice, params.GasCostBase)

		_, tallyResults[i] = k.FilterAndTally(ctx, req, params, gasMeter)
		dataResults[i].Result = tallyResults[i].Result
		dataResults[i].ExitCode = tallyResults[i].ExitCode
		dataResults[i].Consensus = tallyResults[i].Consensus

		k.Logger(ctx).Info("completed tally", "request_id", req.ID)
		k.Logger(ctx).Debug("tally result", "request_id", req.ID, "tally_result", tallyResults[i])

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
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to marshal sudo remove data requests", "err", err)
		return nil
	}
	_, err = k.wasmKeeper.Sudo(ctx, coreContract, msg)
	if err != nil {
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to notify core contract of tally completion", "err", err)
		return nil
	}

	// Store the data results for batching.
	for i := range dataResults {
		err := k.batchingKeeper.SetDataResultForBatching(ctx, dataResults[i])
		// If writing to the store fails we should stop the node to prevent acting on invalid state.
		if err != nil {
			k.Logger(ctx).Error("failed to store data result for batching", "err", err)
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

	telemetry.SetGauge(float32(len(tallyList)), types.TelemetryKeyDataRequestsTallied)

	return nil
}

type TallyResult struct {
	Consensus    bool
	StdOut       []string
	StdErr       []string
	Result       []byte
	ExitCode     uint32
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
	// filterResult, filterErr := ExecuteFilter(reveals, req.ConsensusFilter, req.ReplicationFactor, params, gasMeter)
	outliers := make([]bool, len(reveals))
	for i, _ := range reveals {
		outliers[i] = false
	}
	filterResult := FilterResult{
		Consensus:    true,
		Outliers:     outliers,
		ProxyPubKeys: []string{},
		Errors:       outliers,
	}
	var filterErr error
	tallyResult := TallyResult{
		Consensus:    true,
		ProxyPubKeys: []string{},
	}

	// Phase 2: Tally Program Execution
	var vmRes types.VMResult
	var tallyErr error
	if filterErr == nil {
		vmRes, tallyErr = k.ExecuteTallyProgram(ctx, req, filterResult, reveals, gasMeter)
		if tallyErr != nil {
			tallyResult.Result = []byte(tallyErr.Error())
			tallyResult.ExitCode = types.TallyExitCodeExecError
		} else {
			tallyResult.Result = vmRes.Result
			tallyResult.ExitCode = vmRes.ExitCode
			tallyResult.StdOut = vmRes.Stdout
			tallyResult.StdErr = vmRes.Stderr
		}
	} else {
		tallyResult.Result = []byte(filterErr.Error())
		if errors.Is(filterErr, types.ErrInvalidFilterInput) {
			tallyResult.ExitCode = types.TallyExitCodeInvalidFilterInput
		} else {
			tallyResult.ExitCode = types.TallyExitCodeFilterError
		}
	}

	// Phase 3: Calculate data proxy and executor gas consumption.
	// Calculate data proxy gas consumption if basic consensus was reached.
	if filterErr == nil || !errors.Is(filterErr, types.ErrNoBasicConsensus) {
		k.MeterProxyGas(ctx, tallyResult.ProxyPubKeys, req.ReplicationFactor, gasMeter)
	}

	// Calculate executor gas consumption.
	// switch {
	// case errors.Is(filterErr, types.ErrNoBasicConsensus):
	// 	MeterExecutorGasFallback(req, params.ExecutionGasCostFallback, gasMeter)
	// case errors.Is(filterErr, types.ErrInvalidFilterInput) || errors.Is(filterErr, types.ErrNoConsensus) || tallyErr != nil:
	// 	gasMeter.SetReducedPayoutMode()
	// 	fallthrough
	// default: // filterErr == ErrConsensusInError || filterErr == nil
	// 	gasReports := make([]uint64, len(reveals))
	// 	for i, reveal := range reveals {
	// 		gasReports[i] = reveal.GasUsed
	// 	}
	// 	if areGasReportsUniform(gasReports) {
	// 		MeterExecutorGasUniform(keys, gasReports[0], filterResult.Outliers, req.ReplicationFactor, gasMeter)
	// 	} else {
	// 		MeterExecutorGasDivergent(keys, gasReports, filterResult.Outliers, req.ReplicationFactor, gasMeter)
	// 	}
	// }

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
