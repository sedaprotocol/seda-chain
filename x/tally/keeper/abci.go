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

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func (k Keeper) EndBlock(ctx sdk.Context) (err error) {
	// Use defer to prevent returning an error, which would cause
	// the chain to halt.
	defer func() {
		// Handle a panic.
		if r := recover(); r != nil {
			k.Logger(ctx).Error("recovered from panic in wasm-storage EndBlock", "err", r)
		}
		// Handle an error.
		if err != nil {
			k.Logger(ctx).Error("error in wasm-storage EndBlock", "err", err)
		}
		err = nil
	}()

	err = k.ProcessTallies(ctx)
	if err != nil {
		return
	}
	return
}

// ProcessTallies fetches from the core contract the list of requests
// to be tallied and then goes through it to filter and tally.
func (k Keeper) ProcessTallies(ctx sdk.Context) error {
	// Get core contract address.
	coreContract, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
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
	sudoMsgs := make([]types.Sudo, len(tallyList))
	tallyResults := make([]TallyResult, len(tallyList))
	for i, req := range tallyList {
		// Construct barebone sudo message to be posted to the contract
		// here and populate its results fields after FilterAndTally.
		sudoMsg := types.Sudo{
			ID: req.ID,
			Result: types.DataResult{
				Version:        req.Version,
				ID:             req.ID,
				BlockHeight:    uint64(ctx.BlockHeight()),
				GasUsed:        "0", // TODO
				PaybackAddress: req.PaybackAddress,
				SedaPayload:    req.SedaPayload,
			},
		}

		result, err := k.FilterAndTally(ctx, req)
		if err != nil {
			// Return with exit code 255 to signify that the tally VM
			// was not executed due to the error specified in the result
			// field.
			sudoMsg.ExitCode = 0xff
			sudoMsg.Result.ExitCode = 0xff
			sudoMsg.Result.Result = []byte(err.Error())
			sudoMsg.Result.Consensus = result.consensus
		} else {
			sudoMsg.ExitCode = byte(result.exitInfo.ExitCode)
			sudoMsg.Result.ExitCode = byte(result.exitInfo.ExitCode)
			sudoMsg.Result.Result = result.result
			sudoMsg.Result.Consensus = result.consensus
		}
		k.Logger(ctx).Info(
			"completed tally execution",
			"request_id", req.ID,
			"result", result,
			"sudo_message", sudoMsg,
		)

		sudoMsgs[i] = sudoMsg
		tallyResults[i] = result
	}

	msg, err := json.Marshal(struct {
		PostDataResults struct {
			Results []types.Sudo `json:"results"`
		} `json:"post_data_results"`
	}{
		PostDataResults: struct {
			Results []types.Sudo `json:"results"`
		}{
			Results: sudoMsgs,
		},
	})
	if err != nil {
		return err
	}

	postRes, err := k.wasmKeeper.Sudo(ctx, coreContract, msg)
	if err != nil {
		return err
	}

	for i := range sudoMsgs {
		k.Logger(ctx).Info(
			"tally flow completed",
			"request_id", sudoMsgs[i].ID,
			"post_result", postRes,
		)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTallyCompletion,
				sdk.NewAttribute(types.AttributeDataRequestID, sudoMsgs[i].ID),
				sdk.NewAttribute(types.AttributeTypeConsensus, strconv.FormatBool(sudoMsgs[i].Result.Consensus)),
				sdk.NewAttribute(types.AttributeTallyVMStdOut, strings.Join(tallyResults[i].stdout, "\n")),
				sdk.NewAttribute(types.AttributeTallyVMStdErr, strings.Join(tallyResults[i].stderr, "\n")),
				sdk.NewAttribute(types.AttributeTallyExitCode, fmt.Sprintf("%02x", sudoMsgs[i].ExitCode)),
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

	tallyWasm, err := k.wasmStorageKeeper.GetDataRequestWasm(ctx, req.TallyBinaryID)
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
		"tally_wasm_hash", req.TallyBinaryID,
		"consensus", result.consensus,
		"arguments", args,
	)

	vmRes := tallyvm.ExecuteTallyVm(tallyWasm.Bytecode, args, map[string]string{
		"VM_MODE":               "tally",
		"CONSENSUS":             fmt.Sprintf("%v", result.consensus),
		"DR_ID":                 req.ID,
		"DR_INPUT":              req.DrInputs,
		"BINARY_ID":             req.DrBinaryID,
		"DR_REPLICATION_FACTOR": fmt.Sprintf("%v", req.ReplicationFactor),
		"DR_GAS_PRICE":          req.GasPrice,
		"DR_GAS_LIMIT":          req.GasLimit,
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
