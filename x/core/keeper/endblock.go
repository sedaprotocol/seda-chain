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

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm/v2"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/core/types"
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

	return k.ProcessTallies(ctx, coreContract)
}

// ProcessTallies fetches from the core contract the list of requests
// to be tallied and then goes through it to filter and tally.
func (k Keeper) ProcessTallies(ctx sdk.Context, coreContract sdk.AccAddress) error {
	params, err := k.GetTallyConfig(ctx)
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

	// Loop through the list to apply filter, execute tally, and post
	// execution result.
	processedReqs := make(map[string][]types.Distribution)
	tallyResults := make([]TallyResult, len(tallyList))
	dataResults := make([]batchingtypes.DataResult, len(tallyList))

	for i, req := range tallyList {
		// Initialize the processedReqs map for each request with a full refund (no other distributions)
		processedReqs[req.ID] = make([]types.Distribution, 0)

		dataResults[i], err = req.ToResult(ctx)
		if err != nil {
			markResultErr := types.MarkResultAsFallback(&dataResults[i], err)
			if markResultErr != nil {
				return err
			}
			continue
		}

		if contractQueryResponse.IsPaused {
			markResultErr := types.MarkResultAsPaused(&dataResults[i])
			if markResultErr != nil {
				return err
			}
			continue
		}

		postedGasPrice, ok := math.NewIntFromString(req.PostedGasPrice)
		if !ok || !postedGasPrice.IsPositive() {
			markResultErr := types.MarkResultAsFallback(&dataResults[i], fmt.Errorf("invalid gas price: %s", req.PostedGasPrice))
			if markResultErr != nil {
				return err
			}
			continue
		}

		gasMeter := types.NewGasMeter(req.TallyGasLimit, req.ExecGasLimit, params.MaxTallyGasLimit, postedGasPrice, params.GasCostBase)

		if len(req.Commits) < int(req.ReplicationFactor) {
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", req.ReplicationFactor, len(req.Commits)))
			dataResults[i].ExitCode = types.TallyExitCodeNotEnoughCommits
			k.Logger(ctx).Info("data request's number of commits did not meet replication factor", "request_id", req.ID)

			MeterExecutorGasFallback(req, params.ExecutionGasCostFallback, gasMeter)
		} else {
			_, tallyResults[i] = k.FilterAndTally(ctx, req, params, gasMeter)
			dataResults[i].Result = tallyResults[i].Result
			dataResults[i].ExitCode = tallyResults[i].ExitCode
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
				sdk.NewAttribute(types.AttributeProxyPubKeys, strings.Join(tallyResults[i].ProxyPubKeys, "\n")),
			),
		)
	}

	telemetry.SetGauge(float32(len(tallyList)), types.TelemetryKeyDataRequestsTallied)
	telemetry.SetGauge(0, types.TelemetryKeyDRFlowHalt)

	return nil
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
func (k Keeper) FilterAndTally(ctx sdk.Context, req types.Request, params types.TallyConfig, gasMeter *types.GasMeter) (FilterResult, TallyResult) {
	reveals, executors, gasReports := req.SanitizeReveals(ctx.BlockHeight())

	// Phase 1: Filtering
	filterResult, filterErr := ExecuteFilter(reveals, req.ConsensusFilter, req.ReplicationFactor, params, gasMeter)
	filterResult.Executors = executors
	tallyResult := TallyResult{
		Consensus:    filterResult.Consensus,
		ProxyPubKeys: filterResult.ProxyPubKeys,
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
	switch {
	case errors.Is(filterErr, types.ErrNoBasicConsensus):
		MeterExecutorGasFallback(req, params.ExecutionGasCostFallback, gasMeter)
	case errors.Is(filterErr, types.ErrInvalidFilterInput) || errors.Is(filterErr, types.ErrNoConsensus) || tallyErr != nil:
		gasMeter.SetReducedPayoutMode()
		fallthrough
	default: // filterErr == ErrConsensusInError || filterErr == nil
		if areGasReportsUniform(gasReports) {
			MeterExecutorGasUniform(reveals, gasReports[0], filterResult.Outliers, req.ReplicationFactor, gasMeter)
		} else {
			MeterExecutorGasDivergent(reveals, gasReports, filterResult.Outliers, req.ReplicationFactor, gasMeter)
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
