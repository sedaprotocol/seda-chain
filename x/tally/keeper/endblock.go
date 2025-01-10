package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	processedReqs := make(map[string]types.DistributionMessages)
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

		var distMsgs types.DistributionMessages
		var err error
		switch {
		case len(req.Commits) == 0 || len(req.Commits) < int(req.ReplicationFactor):
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", req.ReplicationFactor, len(req.Commits)))
			dataResults[i].ExitCode = TallyExitCodeNotEnoughCommits
			k.Logger(ctx).Info("data request's number of commits did not meet replication factor", "request_id", req.ID)

			distMsgs, err = k.CalculateCommitterPayouts(ctx, req)
			if err != nil {
				return err
			}
		case len(req.Reveals) == 0 || len(req.Reveals) < int(req.ReplicationFactor):
			dataResults[i].Result = []byte(fmt.Sprintf("need %d reveals; received %d", req.ReplicationFactor, len(req.Reveals)))
			dataResults[i].ExitCode = TallyExitCodeNotEnoughReveals
			k.Logger(ctx).Info("data request's number of reveals did not meet replication factor", "request_id", req.ID)

			distMsgs, err = k.CalculateCommitterPayouts(ctx, req)
			if err != nil {
				return err
			}
		default:
			tallyResults[i] = k.FilterAndTally(ctx, req)
			dataResults[i].Result = tallyResults[i].result
			//nolint:gosec // G115: We shouldn't get negative exit code anyway.
			dataResults[i].ExitCode = uint32(tallyResults[i].exitInfo.ExitCode)
			dataResults[i].Consensus = tallyResults[i].consensus
			dataResults[i].GasUsed = tallyResults[i].execGasUsed + tallyResults[i].tallyGasUsed

			k.Logger(ctx).Info("completed tally", "request_id", req.ID)
			k.Logger(ctx).Debug("tally result", "request_id", req.ID, "tally_result", tallyResults[i])

			// TODO
			distMsgs = types.DistributionMessages{
				Messages:   []types.DistributionMessage{},
				RefundType: types.DistributionTypeNoConsensus,
			}
		}

		processedReqs[req.ID] = distMsgs
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
				sdk.NewAttribute(types.AttributeTallyVMStdOut, strings.Join(tallyResults[i].stdout, "\n")),
				sdk.NewAttribute(types.AttributeTallyVMStdErr, strings.Join(tallyResults[i].stderr, "\n")),
				sdk.NewAttribute(types.AttributeExecGasUsed, fmt.Sprintf("%v", tallyResults[i].execGasUsed)),
				sdk.NewAttribute(types.AttributeTallyGasUsed, fmt.Sprintf("%v", tallyResults[i].tallyGasUsed)),
				sdk.NewAttribute(types.AttributeTallyExitCode, fmt.Sprintf("%02x", dataResults[i].ExitCode)),
				sdk.NewAttribute(types.AttributeProxyPubKeys, strings.Join(tallyResults[i].proxyPubKeys, "\n")),
			),
		)
	}

	return nil
}

type TallyResult struct {
	consensus    bool
	stdout       []string
	stderr       []string
	result       []byte
	exitInfo     tallyvm.ExitInfo
	execGasUsed  uint64
	tallyGasUsed uint64
	proxyPubKeys []string // data proxy pubkeys in basic consensus
}

// FilterAndTally builds and applies filter, executes tally program,
// and calculates payouts.
func (k Keeper) FilterAndTally(ctx sdk.Context, req types.Request) TallyResult {
	var result TallyResult

	// Sort reveals and proxy public keys.
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

	// Phase I: Filtering
	var filterResult FilterResult
	filter, err := k.BuildFilter(ctx, req.ConsensusFilter, req.ReplicationFactor)
	if err != nil {
		result.result = []byte(err.Error())
		result.exitInfo.ExitCode = TallyExitCodeInvalidFilterInput
	} else {
		filterResult, err = ApplyFilter(filter, reveals)
		result.consensus = filterResult.Consensus
		result.proxyPubKeys = filterResult.ProxyPubKeys
		result.tallyGasUsed += filterResult.GasUsed

		// Phase II: Tally Program Execution
		if err != nil {
			result.result = []byte(err.Error())
			result.exitInfo.ExitCode = TallyExitCodeFilterError
		} else {
			vmRes, err := k.ExecuteTallyProgram(ctx, req, filterResult, reveals)
			if err != nil {
				result.result = []byte(err.Error())
				result.exitInfo.ExitCode = TallyExitCodeExecError
			} else {
				result.result = vmRes.Result
				result.exitInfo = vmRes.ExitInfo
				result.stdout = vmRes.Stdout
				result.stderr = vmRes.Stderr
			}
			result.tallyGasUsed += vmRes.GasUsed
		}
	}

	// Phase III: Calculate Payouts
	result.execGasUsed = calculateExecGasUsed(reveals)

	return result
}

// logErrAndRet logs the base error along with the request ID for
// debugging and returns the registered error.
func (k Keeper) logErrAndRet(ctx sdk.Context, baseErr, registeredErr error, req types.Request) error {
	k.Logger(ctx).Debug(baseErr.Error(), "request_id", req.ID, "error", registeredErr)
	return registeredErr
}
