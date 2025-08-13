package app

import (
	"context"
	"errors"

	wasmapp "github.com/CosmWasm/wasmd/app"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
)

// HandlerOptions extends the wasmapp.HandlerOptions with the expected
// Wasm Storage keeper.
type HandlerOptions struct {
	wasmapp.HandlerOptions

	WasmStorageKeeper WasmStorageKeeper
}

type WasmStorageKeeper interface {
	GetCoreContractAddr(ctx context.Context) (sdk.AccAddress, error)
}

// NewAnteHandler wraps the wasmapp.NewAnteHandler with a CommitRevealDecorator.
// We manually prepend the decorator so we don't have to maintain the order of
// the decorators required by CosmWasm.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.WasmStorageKeeper == nil {
		return nil, errors.New("wasm storage keeper is required for ante builder")
	}

	wasmAnteHandler, err := wasmapp.NewAnteHandler(options.HandlerOptions)
	if err != nil {
		return nil, err
	}

	commitRevealDecorator := NewCommitRevealDecorator(options.WasmStorageKeeper)

	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return commitRevealDecorator.AnteHandle(ctx, tx, simulate, wasmAnteHandler)
	}, nil
}

// CommitRevealDecorator guarantees certain properties about transactions
// involving commit/reveal messages.
type CommitRevealDecorator struct {
	wasmStorageKeeper WasmStorageKeeper
}

func NewCommitRevealDecorator(wasmStorageKeeper WasmStorageKeeper) *CommitRevealDecorator {
	return &CommitRevealDecorator{wasmStorageKeeper: wasmStorageKeeper}
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
	commitRevealTx := false
	for i, msg := range tx.GetMsgs() {
		msgInfo, commitRevealMsg := coretypes.ExtractCommitRevealMsgInfo(coreContract.String(), msg)
		// The first message determines the type of the tx.
		if i == 0 {
			commitRevealTx = commitRevealMsg
		}

		// We don't allow mixing commit/reveal messages with other messages.
		if commitRevealTx != commitRevealMsg {
			return sdk.Context{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "commit or reveal message cannot be mixed with other messages")
		}
		// If we don't have a commit/reveal message, we can skip the rest of the loop.
		if !commitRevealTx {
			continue
		}
		// If we see a duplicate commit/reveal message, we return an error.
		if seen[msgInfo] {
			return sdk.Context{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate commit or reveal message detected: %s", msgInfo)
		}
		seen[msgInfo] = true
	}

	return next(ctx, tx, simulate)
}
