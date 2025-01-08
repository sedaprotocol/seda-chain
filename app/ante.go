package app

import (
	"encoding/json"
	"errors"

	wasmapp "github.com/CosmWasm/wasmd/app"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmstoragekeeper "github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
)

// HandlerOptions extends the wasmapp.HandlerOptions with a WasmStorageKeeper.
type HandlerOptions struct {
	wasmapp.HandlerOptions

	WasmStorageKeeper *wasmstoragekeeper.Keeper
}

// NewAnteHandler wraps the wasmapp.NewAnteHandler with a CommitRevealDecorator.
// We manually prepend the decorator so we don't have to maintain the order of the decorators
// required by CosmWasm.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.WasmStorageKeeper == nil {
		return nil, errors.New("wasm storage keeper is required for ante builder")
	}

	wasmAnteHandler, err := wasmapp.NewAnteHandler(options.HandlerOptions)
	if err != nil {
		return nil, err
	}

	commitRevealDecorator := NewCommitRevealDecorator(options.WasmStorageKeeper, options.WasmKeeper)

	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return commitRevealDecorator.AnteHandle(ctx, tx, simulate, wasmAnteHandler)
	}, nil
}

// Decorator which allows for free gas for eligible commit and/or reveal messages.
type CommitRevealDecorator struct {
	wasmStorageKeeper *wasmstoragekeeper.Keeper
	wasmKeeper        *wasmkeeper.Keeper
}

func NewCommitRevealDecorator(wasmStorageKeeper *wasmstoragekeeper.Keeper, wasmKeeper *wasmkeeper.Keeper) *CommitRevealDecorator {
	return &CommitRevealDecorator{wasmStorageKeeper: wasmStorageKeeper, wasmKeeper: wasmKeeper}
}

// AnteHandle checks if a transaction consists of only eligible commit or reveal messages
// and if so, sets the min gas prices to 0.
func (d CommitRevealDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Without a core contract there is no need to check for free gas
	coreContract, err := d.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil || coreContract == nil {
		return next(ctx, tx, simulate)
	}

	// If any message does not qualify for free gas we don't need to check further
	for _, msg := range tx.GetMsgs() {
		if !d.checkFreeGas(ctx, coreContract, msg) {
			return next(ctx, tx, simulate)
		}
	}

	// Only when all messages qualify for free gas, we set the min gas prices to 0
	return next(ctx.WithMinGasPrices(sdk.NewDecCoins()), tx, simulate)
}

// These are the JSON messages used by the overlay when executing the contract, the chain
// uses the same messages to query the contract.
type CommitDataResult struct {
	DrID       string `json:"dr_id"`
	PublicKey  string `json:"public_key"`
	Commitment string `json:"commitment"`
	Proof      string `json:"proof"`
}

type RevealDataResult struct {
	DrID      string `json:"dr_id"`
	PublicKey string `json:"public_key"`
}

type CanExecutorCommitQuery struct {
	CanExecutorCommit CommitDataResult `json:"can_executor_commit"`
}

type CanExecutorRevealQuery struct {
	CanExecutorReveal RevealDataResult `json:"can_executor_reveal"`
}

func (d CommitRevealDecorator) checkFreeGas(ctx sdk.Context, coreContract sdk.AccAddress, msg sdk.Msg) bool {
	switch msg := msg.(type) {
	case *wasmtypes.MsgExecuteContract:
		// Not the core contract, so we don't need to check for free gas
		if msg.Contract != coreContract.String() {
			return false
		}

		contractMsg, err := unmarshalMsg(msg.Msg)
		if err != nil {
			return false
		}

		switch contractMsg := contractMsg.(type) {
		case CommitDataResult:
			result, err := d.queryContract(ctx, coreContract, CanExecutorCommitQuery{CanExecutorCommit: contractMsg})
			if err != nil {
				return false
			}

			return result
		case RevealDataResult:
			result, err := d.queryContract(ctx, coreContract, CanExecutorRevealQuery{CanExecutorReveal: contractMsg})
			if err != nil {
				return false
			}

			return result
		// Not a commit or reveal message, so we don't need to check for free gas
		default:
			return false
		}
	// Not an execute contract message, so we don't need to check for free gas
	default:
		return false
	}
}

func (d CommitRevealDecorator) queryContract(ctx sdk.Context, coreContract sdk.AccAddress, query interface{}) (bool, error) {
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return false, err
	}
	queryRes, err := d.wasmKeeper.QuerySmart(ctx, coreContract, queryBytes)
	if err != nil {
		return false, err
	}

	var result bool
	if err := json.Unmarshal(queryRes, &result); err != nil {
		return false, err
	}

	return result, nil
}

func unmarshalMsg(msg wasmtypes.RawContractMessage) (interface{}, error) {
	// We're only interested in the commit or reveal messages
	var msgData struct {
		CommitDataResult *CommitDataResult `json:"commit_data_result"`
		RevealDataResult *RevealDataResult `json:"reveal_data_result"`
	}
	if err := json.Unmarshal(msg, &msgData); err != nil {
		return nil, err
	}

	if msgData.CommitDataResult != nil {
		return *msgData.CommitDataResult, nil
	}

	if msgData.RevealDataResult != nil {
		return *msgData.RevealDataResult, nil
	}

	return nil, nil
}
