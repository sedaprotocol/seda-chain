package app

import (
	"errors"

	wasmapp "github.com/CosmWasm/wasmd/app"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/utils"
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

// AnteHandle guarantees the following properties about the tx:
// 1. Exclusivity: The tx does not mix commit/reveal messages with other messages.
// 2. Uniqueness: Each of the commit/reveal messages is unique.
func (d CommitRevealDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	coreContract, err := d.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil || coreContract == nil {
		return next(ctx, tx, simulate)
	}

	seen := make(map[string]bool)
	seenCommitReveal := false
	for _, msg := range tx.GetMsgs() {
		msgInfo, commitOrReveal := utils.ExtractCommitRevealMsgInfo(coreContract.String(), msg)
		if commitOrReveal {
			if seen[msgInfo] {
				return sdk.Context{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate commit or reveal message detected: %s", msgInfo)
			}
			seen[msgInfo] = true
			seenCommitReveal = true
		} else if seenCommitReveal {
			return sdk.Context{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "commit or reveal message cannot be mixed with other messages")
		}
	}

	return next(ctx, tx, simulate)
}
