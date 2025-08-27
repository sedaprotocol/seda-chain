package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v3"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

const (
	// MaxDataRequestsPerQuery is the maximum number of data requests that will be retrieved in a single query.
	MaxDataRequestsPerQuery = uint32(50)
)

func (k Keeper) EndBlock(ctx sdk.Context) error {
	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get core contract address", "err", err)
		return nil
	}
	if coreContract == nil {
		k.Logger(ctx).Info("skipping tally end block - core contract has not been registered")
		return nil
	}

	postRes, err := k.wasmKeeper.Sudo(ctx, coreContract, []byte(`{"expire_data_requests":{}}`))
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to expire data requests", "err", err)
		return nil
	}
	k.Logger(ctx).Debug("sudo expire_data_requests", "res", postRes)

	return k.Tally(ctx, coreContract)
}

// Tally fetches from the Core Contract a list of tally-ready requests, tallies
// them, reports results to the contract, and stores results for batching.
func (k Keeper) Tally(ctx sdk.Context, coreContract sdk.AccAddress) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tally params", "err", err)
		return nil
	}
	tallyvm.TallyMaxBytes = uint(params.MaxResultSize)

	contractQueryResponse, err := k.queryContract(ctx, coreContract, params.MaxTalliesPerBlock)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to get tally-ready data requests", "err", err)
		return nil
	}

	tallyList := contractQueryResponse.DataRequests
	if len(tallyList) == 0 {
		k.Logger(ctx).Debug("no tally-ready data requests - skipping tally process")
		return nil
	}
	k.Logger(ctx).Info("non-empty tally list - starting tally process")

	tallyResults, dataResults, processedReqs, err := k.ProcessTallies(ctx, tallyList, params, contractQueryResponse.IsPaused)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to tally data requests", "err", err)
		return nil
	}

	// Notify the Core Contract of tally completion.
	msg, err := types.MarshalSudoRemoveDataRequests(processedReqs)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
		k.Logger(ctx).Error("[HALTS_DR_FLOW] failed to marshal sudo remove data requests", "err", err)
		return nil
	}
	_, err = k.wasmKeeper.Sudo(ctx, coreContract, msg)
	if err != nil {
		telemetry.SetGauge(1, types.TelemetryKeyDRFlowHalt)
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
				sdk.NewAttribute(types.AttributeProxyPubKeys, strings.Join(tallyResults[i].FilterResult.ProxyPubKeys, "\n")),
			),
		)
	}

	telemetry.SetGauge(float32(len(tallyList)), types.TelemetryKeyDataRequestsTallied)
	telemetry.SetGauge(0, types.TelemetryKeyDRFlowHalt)

	return nil
}

// ProcessTallies performs the three phases of the tally process given a list
// of requests: Filtering -> VM execution -> Gas metering and distributions.
// It returns the tally results, data results, processed list of requests
// expected by the Core Contract, and an error.
func (k Keeper) ProcessTallies(ctx sdk.Context, tallyList []types.Request, params types.Params, isPaused bool) ([]types.TallyResult, []batchingtypes.DataResult, map[string][]types.Distribution, error) {
	// tallyResults and dataResults have the same indexing.
	tallyResults := make([]types.TallyResult, len(tallyList))
	dataResults := make([]batchingtypes.DataResult, len(tallyList))

	processedReqs := make(map[string][]types.Distribution)
	tallyExecItems := []TallyParallelExecItem{}

	var err error
	for i, req := range tallyList {
		// Initialize the processedReqs map for each request with a full refund (no other distributions)
		processedReqs[req.ID] = make([]types.Distribution, 0)

		tallyResults[i] = types.TallyResult{
			ID:                req.ID,
			Height:            req.Height,
			ReplicationFactor: req.ReplicationFactor,
		}

		dataResults[i], err = req.ToResult(ctx)
		if err != nil {
			markResultErr := types.MarkResultAsFallback(&dataResults[i], &tallyResults[i], err)
			if markResultErr != nil {
				return nil, nil, nil, err
			}
			continue
		}

		if isPaused {
			markResultErr := types.MarkResultAsPaused(&dataResults[i], &tallyResults[i])
			if markResultErr != nil {
				return nil, nil, nil, err
			}
			continue
		}

		postedGasPrice, ok := math.NewIntFromString(req.PostedGasPrice)
		if !ok || !postedGasPrice.IsPositive() {
			markResultErr := types.MarkResultAsFallback(&dataResults[i], &tallyResults[i], fmt.Errorf("invalid gas price: %s", req.PostedGasPrice))
			if markResultErr != nil {
				return nil, nil, nil, err
			}
			continue
		}

		gasMeter := types.NewGasMeter(req.TallyGasLimit, req.ExecGasLimit, params.MaxTallyGasLimit, postedGasPrice, params.GasCostBase)

		// Phase 1: Filtering
		if len(req.Commits) < int(req.ReplicationFactor) {
			tallyResults[i].FilterResult = types.FilterResult{Error: types.ErrFilterDidNotRun}
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", req.ReplicationFactor, len(req.Commits)))
			dataResults[i].ExitCode = types.TallyExitCodeNotEnoughCommits

			k.Logger(ctx).Info("data request's number of commits did not meet replication factor", "request_id", req.ID)

			MeterExecutorGasFallback(req, params.ExecutionGasCostFallback, gasMeter)
		} else {
			reveals, executors, gasReports := req.SanitizeReveals(ctx.BlockHeight())
			filterResult, filterErr := types.ExecuteFilter(reveals, req.ConsensusFilter, req.ReplicationFactor, params, gasMeter)

			filterResult.Error = filterErr
			filterResult.Executors = executors

			tallyResults[i].Reveals = reveals
			tallyResults[i].GasReports = gasReports
			tallyResults[i].FilterResult = filterResult
			dataResults[i].Consensus = filterResult.Consensus

			if filterErr == nil {
				// Execute tally VM for this request.
				tallyExecItems = append(tallyExecItems, NewTallyParallelExecItem(i, req, gasMeter, reveals, filterResult.Outliers, filterResult.Consensus))
			} else {
				// Skip tally execution.
				dataResults[i].Result = []byte(filterErr.Error())
				if errors.Is(filterErr, types.ErrInvalidFilterInput) {
					dataResults[i].ExitCode = types.TallyExitCodeInvalidFilterInput
				} else {
					dataResults[i].ExitCode = types.TallyExitCodeFilterError
				}

				if errors.Is(filterErr, types.ErrNoBasicConsensus) {
					MeterExecutorGasFallback(req, params.ExecutionGasCostFallback, gasMeter)
				} else if errors.Is(filterErr, types.ErrInvalidFilterInput) || errors.Is(filterErr, types.ErrNoConsensus) {
					gasMeter.SetReducedPayoutMode()
				}
			}
		}

		tallyResults[i].GasMeter = gasMeter
	}

	// Phase 2: Parallel execution of tally VM
	if len(tallyExecItems) > 0 {
		vmResults := k.ExecuteTallyProgramsParallel(ctx, tallyExecItems)

		// Populate tallyResults and dataResults with the results of the execution.
		vmResultIndex := 0
		for i := range tallyExecItems {
			if tallyExecItems[i].TallyExecErr == nil {
				tallyExecItems[i].GasMeter.ConsumeTallyGas(vmResults[vmResultIndex].GasUsed)

				result := types.MapVMResult(vmResults[vmResultIndex])
				if result.ExitCode != 0 {
					k.Logger(ctx).Error("tally vm exit message", "request_id", tallyExecItems[i].Request.ID, "exit_message", result.ExitMessage)
				}

				resultIndex := tallyExecItems[i].Index
				tallyResults[resultIndex].StdOut = result.Stdout
				tallyResults[resultIndex].StdErr = result.Stderr
				dataResults[resultIndex].Result = result.Result
				dataResults[resultIndex].ExitCode = result.ExitCode
				vmResultIndex++
			} else {
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
			k.MeterProxyGas(ctx, tr.FilterResult.ProxyPubKeys, tr.ReplicationFactor, tr.GasMeter)

			if areGasReportsUniform(tr.GasReports) {
				tr.MeterExecutorGasUniform()
			} else {
				tr.MeterExecutorGasDivergent()
			}
		}

		// GasMeter may not have been initialized in some cases.
		if tr.GasMeter != nil {
			tallyResults[i].TallyGasUsed = tr.GasMeter.TallyGasUsed()
			tallyResults[i].ExecGasUsed = tr.GasMeter.ExecutionGasUsed()

			processedReqs[tr.ID] = k.DistributionsFromGasMeter(ctx, tr.ID, tr.Height, tr.GasMeter, params.BurnRatio)
			dataResults[i].GasUsed = tr.GasMeter.TotalGasUsed()
		}

		dataResults[i].Id, err = dataResults[i].TryHash()
		if err != nil {
			return nil, nil, nil, err
		}

		k.Logger(ctx).Info("completed tally", "request_id", tr.ID)
		k.Logger(ctx).Debug("tally result", "request_id", tr.ID, "tally_result", tr)
	}

	return tallyResults, dataResults, processedReqs, nil
}

// queryContract fetches tally-ready data requests from the core contract in batches while
// keeping the interface consistent. This avoids problems where the contract runs out of memory
// when we fetch the entire maxTalliesPerBlock in a single query.
func (k Keeper) queryContract(ctx sdk.Context, coreContract sdk.AccAddress, maxTalliesPerBlock uint32) (*types.ContractListResponse, error) {
	tallyList := make([]types.Request, 0, maxTalliesPerBlock)
	lastSeenIndex := types.EmptyLastSeenIndex()
	isPaused := false

	for {
		// The limit is the smaller value between the max number of data requests per query or
		// the remaining number that still fits in the block tally limit.
		//nolint:gosec // G115: the length of a list should never be negative.
		limit := min(MaxDataRequestsPerQuery, maxTalliesPerBlock-uint32(len(tallyList)))

		// Fetch tally-ready data requests.
		queryRes, err := k.wasmViewKeeper.QuerySmart(
			ctx, coreContract,
			fmt.Appendf(nil, `{"get_data_requests_by_status":{"status": "tallying", "last_seen_index": %s, "limit": %d}}`, lastSeenIndex.String(), limit),
		)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to query contract")
		}

		var contractQueryResponse types.ContractListResponse
		err = json.Unmarshal(queryRes, &contractQueryResponse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to unmarshal data requests contract response")
		}

		lastSeenIndex = contractQueryResponse.LastSeenIndex
		isPaused = contractQueryResponse.IsPaused
		tallyList = append(tallyList, contractQueryResponse.DataRequests...)

		// Break if we've reached the max number of data requests or if the
		// number of data requests returned is less than the limit.
		if len(tallyList) >= int(maxTalliesPerBlock) || len(contractQueryResponse.DataRequests) < int(limit) {
			break
		}
	}

	return &types.ContractListResponse{
		IsPaused:     isPaused,
		DataRequests: tallyList,
	}, nil
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
