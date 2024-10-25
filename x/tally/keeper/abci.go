package keeper

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-wasm-vm/tallyvm"

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
	tallyResults := make([]TallyResult, len(tallyList))
	sudoMsgs := make([]types.SudoRemoveDataRequest, len(tallyList))
	dataResults := make([]batchingtypes.DataResult, len(tallyList))
	for i, req := range tallyList {
		dataResults[i] = batchingtypes.DataResult{
			DrId:    req.ID,
			Version: req.Version,
			//nolint:gosec // G115: We shouldn't get negative block heights anyway.
			BlockHeight:    uint64(ctx.BlockHeight()),
			GasUsed:        0, // TODO (#367)
			PaybackAddress: req.PaybackAddress,
			SedaPayload:    req.SedaPayload,
		}

		// Check for expiration.
		if len(req.Commits) < int(req.ReplicationFactor) {
			dataResults[i].Result = []byte(fmt.Sprintf("need %d commits; received %d", req.ReplicationFactor, len(req.Commits)))
			dataResults[i].ExitCode = 200
			continue
		}
		if len(req.Reveals) < int(req.ReplicationFactor) {
			dataResults[i].Result = []byte(fmt.Sprintf("need %d reveals; received %d", req.ReplicationFactor, len(req.Commits)))
			dataResults[i].ExitCode = 201
			continue
		}

		result, err := k.FilterAndTally(ctx, req)
		if err != nil {
			// Return with exit code 255 to signify that the tally VM
			// was not executed due to the error specified in the result
			// field.
			dataResults[i].ExitCode = 0xff
			dataResults[i].Result = []byte(err.Error())
			dataResults[i].Consensus = result.consensus
		} else {
			//nolint:gosec // G115: We shouldn't get negative exit code anyway.
			dataResults[i].ExitCode = uint32(result.exitInfo.ExitCode)
			dataResults[i].Result = result.result
			dataResults[i].Consensus = result.consensus
		}
		k.Logger(ctx).Info(
			"completed tally execution",
			"request_id", req.ID,
			"tally_result", result,
		)

		dataResults[i].Id, err = dataResults[i].TryHash()
		if err != nil {
			return err
		}
		sudoMsgs[i] = types.SudoRemoveDataRequest{ID: req.ID}
		tallyResults[i] = result
	}

	// Notify the Core Contract of tally completion.
	msg, err := json.Marshal(struct {
		SudoRemoveDataRequests struct {
			Requests []types.SudoRemoveDataRequest `json:"requests"`
		} `json:"remove_data_requests"`
	}{
		SudoRemoveDataRequests: struct {
			Requests []types.SudoRemoveDataRequest `json:"requests"`
		}{
			Requests: sudoMsgs,
		},
	})
	if err != nil {
		return err
	}
	postRes, err := k.wasmKeeper.Sudo(ctx, coreContract, msg)
	if err != nil {
		return err
	}

	// Store the data results for batching.
	for i := range dataResults {
		err := k.batchingKeeper.SetDataResultForBatching(ctx, dataResults[i])
		if err != nil {
			return err
		}
	}

	for i := range sudoMsgs {
		k.Logger(ctx).Info(
			"tally flow completed",
			"request_id", dataResults[i].DrId,
			"post_result", postRes,
		)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTallyCompletion,
				sdk.NewAttribute(types.AttributeDataResultID, dataResults[i].Id),
				sdk.NewAttribute(types.AttributeDataRequestID, dataResults[i].DrId),
				sdk.NewAttribute(types.AttributeTypeConsensus, strconv.FormatBool(dataResults[i].Consensus)),
				sdk.NewAttribute(types.AttributeTallyVMStdOut, strings.Join(tallyResults[i].stdout, "\n")),
				sdk.NewAttribute(types.AttributeTallyVMStdErr, strings.Join(tallyResults[i].stderr, "\n")),
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
	proxyPubKeys []string // data proxy pubkeys in basic consensus
}

// FilterAndTally applies filter and executes tally. It returns the
// tally VM result, consensus boolean, consensus data proxy public keys,
// and error if applicable.
func (k Keeper) FilterAndTally(ctx sdk.Context, req types.Request) (TallyResult, error) {
	var result TallyResult
	filter, err := base64.StdEncoding.DecodeString(req.ConsensusFilter)
	if err != nil {
		return result, errorsmod.Wrap(err, "failed to decode consensus filter")
	}
	// Convert base64-encoded payback address to hex encoding that
	// the tally VM expects.
	decodedBytes, err := base64.StdEncoding.DecodeString(req.PaybackAddress)
	if err != nil {
		return result, errorsmod.Wrap(err, "failed to decode payback address")
	}
	paybackAddrHex := hex.EncodeToString(decodedBytes)

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

	var outliers []int
	outliers, result.consensus, result.proxyPubKeys, err = ApplyFilter(filter, reveals)
	if err != nil {
		return result, errorsmod.Wrap(err, "error while applying filter")
	}

	tallyProgram, err := k.wasmStorageKeeper.GetOracleProgram(ctx, req.TallyProgramID)
	if err != nil {
		return result, err
	}
	tallyInputs, err := base64.StdEncoding.DecodeString(req.TallyInputs)
	if err != nil {
		return result, errorsmod.Wrap(err, "failed to decode tally inputs")
	}

	args, err := tallyVMArg(tallyInputs, reveals, outliers)
	if err != nil {
		return result, errorsmod.Wrap(err, "failed to construct tally VM arguments")
	}

	k.Logger(ctx).Info(
		"executing tally VM",
		"request_id", req.ID,
		"tally_program_id", req.TallyProgramID,
		"consensus", result.consensus,
		"arguments", args,
	)

	vmRes := tallyvm.ExecuteTallyVm(tallyProgram.Bytecode, args, map[string]string{
		"VM_MODE":               "tally",
		"CONSENSUS":             fmt.Sprintf("%v", result.consensus),
		"DR_ID":                 req.ID,
		"DR_INPUT":              req.DrInputs,
		"BINARY_ID":             req.DrBinaryID,
		"DR_REPLICATION_FACTOR": fmt.Sprintf("%v", req.ReplicationFactor),
		"DR_GAS_PRICE":          req.GasPrice,
		"DR_GAS_LIMIT":          fmt.Sprintf("%v", req.GasLimit),
		"DR_MEMO":               req.Memo,
		"DR_PAYBACK_ADDRESS":    paybackAddrHex,
		"BLOCK_HEIGHT":          fmt.Sprintf("%d", ctx.BlockHeight()),
	})
	result.stdout = vmRes.Stdout
	result.stderr = vmRes.Stderr
	result.result = vmRes.Result
	result.exitInfo = vmRes.ExitInfo

	return result, nil
}

func tallyVMArg(inputArgs []byte, reveals []types.RevealBody, outliers []int) ([]string, error) {
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
