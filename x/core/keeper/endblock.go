package keeper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v3"

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

	return k.Tally(ctx)
}

// Tally fetches from a list of tally-ready requests, tallies them, reports
// results to the contract, and stores results for batching.
func (k Keeper) Tally(ctx sdk.Context) error {
	drIDs, err := k.GetDataRequestIDsByStatus(ctx, types.DATA_REQUEST_STATUS_TALLYING)
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

	params, err := k.GetParams(ctx)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tally params", "err", err)
		return nil
	}
	tallyvm.TallyMaxBytes = uint(params.TallyConfig.MaxResultSize)

	tallyResults, dataResults, err := k.ProcessTallies(ctx, drIDs, params.TallyConfig, params.StakingConfig.MinimumStake)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to tally data requests", "err", err)
		return nil
	}

	// TODO remove_requests.rs

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
				sdk.NewAttribute(types.AttributeProxyPubKeys, strings.Join(tallyResults[i].FilterResult.ProxyPubKeys, "\n")),
			),
		)
	}

	telemetry.SetGauge(float32(tallyLen), types.TelemetryKeyDataRequestsTallied)
	telemetry.SetGauge(0, types.TelemetryKeyDRFlowHalt)

	return nil
}

// ProcessTallies performs the three phases of the tally process given a list
// of requests: Filtering -> VM execution -> Gas metering and distributions.
// It returns the tally results, data results, processed list of requests
// expected by the Core Contract, and an error.
func (k Keeper) ProcessTallies(ctx sdk.Context, drIDs []string, config types.TallyConfig, minStake math.Int) ([]TallyResult, []batchingtypes.DataResult, error) {
	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, nil, err
	}

	// tallyResults and dataResults have the same indexing.
	tallyResults := make([]TallyResult, len(drIDs))
	dataResults := make([]batchingtypes.DataResult, len(drIDs))

	tallyExecItems := []TallyParallelExecItem{}

	for i, id := range drIDs {
		dr, err := k.GetDataRequest(ctx, id)
		if err != nil {
			telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
			k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to retrieve data request", "err", err)
			return nil, nil, err
		}

		dataResults[i] = batchingtypes.DataResult{
			DrId: dr.ID,
			//nolint:gosec // G115: Block height is never negative.
			DrBlockHeight: uint64(dr.PostedHeight),
			Version:       dr.Version,
			//nolint:gosec // G115: Block height is never negative.
			BlockHeight: uint64(ctx.BlockHeight()),
			//nolint:gosec // G115: Timestamp is never negative.
			BlockTimestamp: uint64(ctx.BlockTime().Unix()),
		}

		isPaused, err := k.IsPaused(ctx)
		if err != nil {
			telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
			k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get paused status", "err", err)
			return nil, nil, err
		}

		if isPaused {
			markResultErr := MarkResultAsPaused(&dataResults[i], &tallyResults[i])
			if markResultErr != nil {
				return nil, nil, err
			}
			poster, err := sdk.AccAddressFromBech32(dr.Poster)
			if err != nil {
				// should never happen as the address was validated on posting
				return nil, nil, err
			}
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, poster, sdk.NewCoins(sdk.NewCoin(denom, dr.Escrow)))
			if err != nil {
				return nil, nil, err
			}

			err = k.RemoveRevealBodies(ctx, dr.ID)
			if err != nil {
				return nil, nil, err
			}
			err = k.RemoveDataRequest(ctx, dr.Index(), dr.Status)
			if err != nil {
				return nil, nil, err
			}

			continue
		}

		tallyResults[i] = TallyResult{
			ID: dr.ID,
			//nolint:gosec // G115: Block height is never negative.
			Height: uint64(dr.PostedHeight),
			//nolint:gosec // G115: Replication factor is guaranteed to fit within uint16.
			ReplicationFactor: uint16(dr.ReplicationFactor),
			GasMeter:          types.NewGasMeter(&dr, config.MaxTallyGasLimit, config.BaseGasCost),
		}

		// Phase 1: Filtering
		if len(dr.Commits) < int(dr.ReplicationFactor) {
			tallyResults[i].FilterResult = FilterResult{Error: types.ErrFilterDidNotRun}
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", dr.ReplicationFactor, len(dr.Commits)))
			dataResults[i].ExitCode = types.TallyExitCodeNotEnoughCommits

			k.Logger(ctx).Info("data request's number of commits did not meet replication factor", "request_id", dr.ID)

			MeterExecutorGasFallback(dr, config.ExecutionGasCostFallback, tallyResults[i].GasMeter)
		} else {
			reveals, executors, gasReports := k.LoadRevealsHashSorted(ctx, dr.ID, dr.Reveals, types.GetEntropy(dr.ID, ctx.BlockHeight()))
			//nolint:gosec // G115: Replication factor is guaranteed to fit within uint16.
			filterResult, filterErr := ExecuteFilter(reveals, dr.ConsensusFilter, uint16(dr.ReplicationFactor), config, tallyResults[i].GasMeter)

			filterResult.Error = filterErr
			filterResult.Executors = executors

			tallyResults[i].Reveals = reveals
			tallyResults[i].GasReports = gasReports
			tallyResults[i].FilterResult = filterResult
			dataResults[i].Consensus = filterResult.Consensus

			if filterErr == nil {
				// Execute tally VM for this request.
				tallyExecItems = append(
					tallyExecItems,
					NewTallyParallelExecItem(i, dr, tallyResults[i].GasMeter.RemainingTallyGas(), reveals, filterResult.Outliers, filterResult.Consensus),
				)
			} else {
				// Skip tally execution.
				dataResults[i].Result = []byte(filterErr.Error())
				if errors.Is(filterErr, types.ErrInvalidFilterInput) {
					dataResults[i].ExitCode = types.TallyExitCodeInvalidFilterInput
				} else {
					dataResults[i].ExitCode = types.TallyExitCodeFilterError
				}

				if errors.Is(filterErr, types.ErrNoBasicConsensus) {
					MeterExecutorGasFallback(dr, config.ExecutionGasCostFallback, tallyResults[i].GasMeter)
				} else if errors.Is(filterErr, types.ErrInvalidFilterInput) || errors.Is(filterErr, types.ErrNoConsensus) {
					tallyResults[i].GasMeter.SetReducedPayoutMode()
				}
			}
		}

		err = k.RemoveRevealBodies(ctx, dr.ID)
		if err != nil {
			return nil, nil, err
		}
		err = k.RemoveDataRequest(ctx, dr.Index(), dr.Status)
		if err != nil {
			return nil, nil, err
		}
	}

	// Phase 2: Parallel execution of tally VM
	if len(tallyExecItems) > 0 {
		vmResults := k.ExecuteTallyProgramsParallel(ctx, tallyExecItems)

		// Populate tallyResults and dataResults with the results of the execution.
		vmResultIndex := 0
		for i := range tallyExecItems {
			if tallyExecItems[i].TallyExecErr == nil {
				// Tally was executed, so parse VM execution results.
				result := types.MapVMResult(vmResults[vmResultIndex])
				if result.ExitCode != 0 {
					k.Logger(ctx).Error("tally vm exit message", "request_id", tallyExecItems[i].Request.ID, "exit_message", result.ExitMessage)
				}

				resultIndex := tallyExecItems[i].Index
				tallyResults[resultIndex].StdOut = result.Stdout
				tallyResults[resultIndex].StdErr = result.Stderr
				tallyResults[resultIndex].GasMeter.ConsumeTallyGas(vmResults[vmResultIndex].GasUsed)

				dataResults[resultIndex].Result = result.Result
				dataResults[resultIndex].ExitCode = result.ExitCode
				vmResultIndex++
			} else {
				// Tally was not executed.
				resultIndex := tallyExecItems[i].Index
				tallyResults[resultIndex].GasMeter.SetReducedPayoutMode()
				dataResults[resultIndex].Result = []byte(tallyExecItems[i].TallyExecErr.Error())
				dataResults[resultIndex].ExitCode = types.TallyExitCodeExecError
			}
		}
	}

	// Phase 3: Gas metering and distributions
	for i, tr := range tallyResults {
		filterErr := tr.FilterResult.Error

		// Calculate data proxy and executor gas consumptions if basic consensus
		// was reached.
		if !errors.Is(filterErr, types.ErrNoBasicConsensus) && !errors.Is(filterErr, types.ErrFilterDidNotRun) {
			k.MeterProxyGas(ctx, tr.FilterResult.ProxyPubKeys, uint64(tr.ReplicationFactor), tr.GasMeter)

			if areGasReportsUniform(tr.GasReports) {
				tr.MeterExecutorGasUniform()
			} else {
				tr.MeterExecutorGasDivergent()
			}
		}

		// GasMeter is not initialized under paused cases.
		if tr.GasMeter != nil {
			tallyResults[i].TallyGasUsed = tr.GasMeter.TallyGasUsed()
			tallyResults[i].ExecGasUsed = tr.GasMeter.ExecutionGasUsed()

			err = k.ChargeGasCosts(ctx, denom, &tr, minStake, config.BurnRatio)
			if err != nil {
				return nil, nil, err
			}

			dataResults[i].GasUsed = tr.GasMeter.TotalGasUsed()
		}

		dataResults[i].Id, err = dataResults[i].TryHash()
		if err != nil {
			return nil, nil, err
		}

		k.Logger(ctx).Info("completed tally", "request_id", tr.ID)
		k.Logger(ctx).Debug("tally result", "request_id", tr.ID, "tally_result", tr)
	}

	return tallyResults, dataResults, nil
}

// areGasReportsUniform returns true if the gas reports of the given reveals are
// uniform.
func areGasReportsUniform(reports []uint64) bool {
	if len(reports) <= 1 {
		return true
	}
	firstGas := reports[0]
	for i := 1; i < len(reports); i++ {
		if reports[i] != firstGas {
			return false
		}
	}
	return true
}
