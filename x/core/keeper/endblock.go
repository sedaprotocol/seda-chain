package keeper

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

const (
	// MaxDataRequestsPerQuery is the maximum number of data requests that will be retrieved in a single query.
	MaxDataRequestsPerQuery = uint32(50)
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	// TODO Memory considerations (Check old queryContract with params.MaxTalliesPerBlock)
	err := k.ExpireDataRequests(ctx)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to expire data requests", "err", err)
		return nil
	}

	return k.ProcessTallies(ctx)
}

func (k Keeper) ProcessTallies(ctx sdk.Context) error {
	drIDs, err := k.GetTallyingDataRequestIDs(ctx)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tallying data request IDs", "err", err)
		return nil
	}

	tallyLen := len(drIDs)
	if tallyLen == 0 {
		k.Logger(ctx).Debug("no tally-ready data requests - skipping tally process")
		return nil
	}
	k.Logger(ctx).Info("non-empty tally list - starting tally process")

	params, err := k.GetTallyConfig(ctx)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tally params", "err", err)
		return nil
	}
	tallyvm.TallyMaxBytes = uint(params.MaxResultSize)

	// Loop through the list to apply filter, execute tally, and post
	// execution result.
	processedReqs := make(map[string][]types.Distribution)
	tallyResults := make([]TallyResult, tallyLen)
	dataResults := make([]batchingtypes.DataResult, tallyLen)

	for i, id := range drIDs {
		dr, err := k.DataRequests.Get(ctx, id)
		if err != nil {
			telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
			k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to retrieve data request", "err", err)
			return nil
		}

		// Initialize the processedReqs map for each request with a full refund (no other distributions)
		processedReqs[dr.Id] = make([]types.Distribution, 0)

		dataResults[i] = batchingtypes.DataResult{
			DrId:          dr.Id,
			DrBlockHeight: dr.Height,
			Version:       dr.Version,
			//nolint:gosec // G115: We shouldn't get negative block heights.
			BlockHeight: uint64(ctx.BlockHeight()),
			//nolint:gosec // G115: We shouldn't get negative timestamps.
			BlockTimestamp: uint64(ctx.BlockTime().Unix()),
		}

		// TODO Add pausability
		// if contractQueryResponse.IsPaused {
		// 	markResultErr := types.MarkResultAsPaused(&dataResults[i])
		// 	if markResultErr != nil {
		// 		return err
		// 	}
		// 	continue
		// }

		gasMeter := types.NewGasMeter(dr.TallyGasLimit, dr.ExecGasLimit, params.MaxTallyGasLimit, dr.PostedGasPrice, params.GasCostBase)

		if len(dr.Commits) < int(dr.ReplicationFactor) {
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", dr.ReplicationFactor, len(dr.Commits)))
			dataResults[i].ExitCode = types.TallyExitCodeNotEnoughCommits
			k.Logger(ctx).Info("data request's number of commits did not meet replication factor", "request_id", dr.Id)

			MeterExecutorGasFallback(dr, params.ExecutionGasCostFallback, gasMeter)
		} else {
			_, tallyResults[i] = k.FilterAndTally(ctx, dr, params, gasMeter)
			dataResults[i].Result = tallyResults[i].Result
			dataResults[i].ExitCode = tallyResults[i].ExitCode
			dataResults[i].Consensus = tallyResults[i].Consensus

			k.Logger(ctx).Info("completed tally", "request_id", dr.Id)
			k.Logger(ctx).Debug("tally result", "request_id", dr.Id, "tally_result", tallyResults[i])
		}

		processedReqs[dr.Id] = k.DistributionsFromGasMeter(ctx, dr.Id, dr.Height, gasMeter, params.BurnRatio)

		dataResults[i].GasUsed = gasMeter.TotalGasUsed()
		dataResults[i].Id, err = dataResults[i].TryHash()
		if err != nil {
			return err
		}
	}

	// TODO remove_requests.rs

	// // Notify the Core Contract of tally completion.
	// msg, err := types.MarshalSudoRemoveDataRequests(processedReqs)
	// if err != nil {
	// 	telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
	// 	k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to marshal sudo remove data requests", "err", err)
	// 	return nil
	// }
	// _, err = k.wasmKeeper.Sudo(ctx, coreContract, msg)
	// if err != nil {
	// 	telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
	// 	k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to notify core contract of tally completion", "err", err)
	// 	return nil
	// }

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

	telemetry.SetGauge(float32(tallyLen), types.TelemetryKeyDataRequestsTallied)
	telemetry.SetGauge(0, types.TelemetryKeyDRFlowHalt)

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
func (k Keeper) FilterAndTally(ctx sdk.Context, dr types.DataRequest, params types.TallyConfig, gasMeter *types.GasMeter) (FilterResult, TallyResult) {
	// reveals, executors, gasReports := k.LoadRevealsSorted(ctx, dr.Id, dr.Reveals)
	unsortedReveals := make([]types.Reveal, len(dr.Reveals))
	i := 0
	for executor, revealBody := range dr.Reveals {
		unsortedReveals[i] = types.Reveal{Executor: executor, RevealBody: *revealBody}
		sort.Strings(unsortedReveals[i].ProxyPubKeys)
		i++
	}

	reveals := types.HashSort(unsortedReveals, types.GetEntropy(dr.Id, ctx.BlockHeight()))

	executors := make([]string, len(reveals))
	gasReports := make([]uint64, len(reveals))
	for i, reveal := range reveals {
		executors[i] = reveal.Executor
		gasReports[i] = reveal.GasUsed
	}

	// Phase 1: Filtering
	filterResult, filterErr := ExecuteFilter(reveals, dr.ConsensusFilter, uint16(dr.ReplicationFactor), params, gasMeter)
	filterResult.Executors = executors
	tallyResult := TallyResult{
		Consensus:    filterResult.Consensus,
		ProxyPubKeys: filterResult.ProxyPubKeys,
	}

	// Phase 2: Tally Program Execution
	var vmRes types.VMResult
	var tallyErr error
	if filterErr == nil {
		vmRes, tallyErr = k.ExecuteTallyProgram(ctx, dr, filterResult, reveals, gasMeter)
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
		k.MeterProxyGas(ctx, tallyResult.ProxyPubKeys, uint64(dr.ReplicationFactor), gasMeter)
	}

	// Calculate executor gas consumption.
	switch {
	case errors.Is(filterErr, types.ErrNoBasicConsensus):
		MeterExecutorGasFallback(dr, params.ExecutionGasCostFallback, gasMeter)
	case errors.Is(filterErr, types.ErrInvalidFilterInput) || errors.Is(filterErr, types.ErrNoConsensus) || tallyErr != nil:
		gasMeter.SetReducedPayoutMode()
		fallthrough
	default: // filterErr == ErrConsensusInError || filterErr == nil
		if areGasReportsUniform(gasReports) {
			MeterExecutorGasUniform(reveals, gasReports[0], filterResult.Outliers, uint64(dr.ReplicationFactor), gasMeter)
		} else {
			MeterExecutorGasDivergent(reveals, gasReports, filterResult.Outliers, uint64(dr.ReplicationFactor), gasMeter)
		}
	}

	tallyResult.TallyGasUsed = gasMeter.TallyGasUsed()
	tallyResult.ExecGasUsed = gasMeter.ExecutionGasUsed()
	return filterResult, tallyResult
}

// logErrAndRet logs the base error along with the request ID for
// debugging and returns the registered error.
func (k Keeper) logErrAndRet(ctx sdk.Context, baseErr, registeredErr error, drID string) error {
	k.Logger(ctx).Debug(baseErr.Error(), "request_id", drID, "error", registeredErr)
	return registeredErr
}
