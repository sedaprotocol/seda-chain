package keeper

import (
	"context"
	"encoding/hex"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/prometheus/client_golang/prometheus"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

var OracleProgramCountMetric = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "seda_oracle_program_count",
	Help: "The count of stored oracle programs",
})

// StoreOracleProgram stores an oracle program. It unzips a gzip-
// compressed wasm and stores it using its hash as the key. If a
// duplicate wasm already exists, an error is returned.
func (m msgServer) StoreOracleProgram(goCtx context.Context, msg *types.MsgStoreOracleProgram) (*types.MsgStoreOracleProgramResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	params, err := m.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	unzipped, err := unzipWasm(msg.Wasm, params.MaxWasmSize)
	if err != nil {
		return nil, err
	}

	program := types.NewOracleProgram(unzipped, ctx.BlockTime(), ctx.BlockHeight(), params.WasmTTL)
	if exists, _ := m.OracleProgram.Has(ctx, program.Hash); exists {
		return nil, types.ErrWasmAlreadyExists
	}
	if err := m.OracleProgram.Set(ctx, program.Hash, program); err != nil {
		return nil, err
	}

	if err := m.OracleProgramExpiration.Set(ctx, collections.Join(program.ExpirationHeight, program.Hash)); err != nil {
		return nil, err
	}

	hashHex := hex.EncodeToString(program.Hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStoreOracleProgram,
			sdk.NewAttribute(types.AttributeSender, msg.Sender),
			sdk.NewAttribute(types.AttributeOracleProgramHash, hashHex),
		),
	)

	if !ctx.IsCheckTx() {
		// only increment the metric if the transaction is included in a block
		OracleProgramCountMetric.Inc()
	}

	return &types.MsgStoreOracleProgramResponse{
		Hash: hashHex,
	}, nil
}

// StoreOracleProgram stores an executor wasm used in the SEDA protocol.
// It unzips a gzip-compressed wasm and stores it using its hash as the key.
// If a duplicate wasm already exists, an error is returned.
func (m msgServer) StoreExecutorWasm(goCtx context.Context, msg *types.MsgStoreExecutorWasm) (*types.MsgStoreExecutorWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Sender != m.authority {
		return nil, types.ErrInvalidAuthority.Wrapf("expected %s, got %s", m.authority, msg.Sender)
	}

	params, err := m.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	unzipped, err := unzipWasm(msg.Wasm, params.MaxWasmSize)
	if err != nil {
		return nil, err
	}

	wasm := types.NewExecutorWasm(unzipped, ctx.BlockTime())
	exists, _ := m.Keeper.ExecutorWasm.Has(ctx, wasm.Hash)
	if exists {
		return nil, types.ErrWasmAlreadyExists
	}
	if err = m.Keeper.ExecutorWasm.Set(ctx, wasm.Hash, wasm); err != nil {
		return nil, err
	}

	hashString := hex.EncodeToString(wasm.Hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorWasm,
			sdk.NewAttribute(types.AttributeSender, msg.Sender),
			sdk.NewAttribute(types.AttributeExecutorWasmHash, hashString),
		),
	)

	return &types.MsgStoreExecutorWasmResponse{
		Hash: hashString,
	}, nil
}

// InstantiateCoreContract instantiate a new contract with a
// predictable address and updates the core contract registry.
func (m msgServer) InstantiateCoreContract(goCtx context.Context, msg *types.MsgInstantiateCoreContract) (*types.MsgInstantiateCoreContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	var adminAddr sdk.AccAddress
	var err error
	if msg.Admin != "" {
		if adminAddr, err = sdk.AccAddressFromBech32(msg.Admin); err != nil {
			return nil, err
		}
	}

	contractAddr, _, err := m.wasmKeeper.Instantiate2(ctx, msg.CodeID, adminAddr, adminAddr, msg.Msg, msg.Label, msg.Funds, msg.Salt, msg.FixMsg)
	if err != nil {
		return nil, err
	}

	// Update the core contract registry.
	err = m.CoreContractRegistry.Set(ctx, contractAddr.String())
	if err != nil {
		return nil, err
	}

	return &types.MsgInstantiateCoreContractResponse{
		ContractAddress: contractAddr.String(),
	}, nil
}

// unzipWasm unzips a gzipped wasm.
func unzipWasm(wasm []byte, maxSize int64) ([]byte, error) {
	var unzipped []byte
	var err error
	if !ioutils.IsGzip(wasm) {
		return nil, types.ErrWasmNotGzipCompressed
	}
	unzipped, err = ioutils.Uncompress(wasm, maxSize)
	if err != nil {
		return nil, err
	}
	return unzipped, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority.
	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, err
	}
	if m.GetAuthority() != req.Authority {
		return nil, types.ErrInvalidAuthority.Wrapf("expected %s, got %s", m.authority, req.Authority)
	}

	// Validate and update module parameters.
	if err := req.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.Params.Set(ctx, req.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
