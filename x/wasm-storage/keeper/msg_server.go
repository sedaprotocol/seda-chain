package keeper

import (
	"context"
	"encoding/hex"
	"math"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/prometheus/client_golang/prometheus"

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

	program := types.NewOracleProgram(unzipped, ctx.BlockTime())
	if exists, _ := m.OracleProgram.Has(ctx, program.Hash); exists {
		return nil, types.ErrWasmAlreadyExists
	}

	beforeGas := ctx.GasMeter().GasConsumed()
	if err := m.OracleProgram.Set(ctx, program.Hash, program); err != nil {
		return nil, err
	}

	// Apply the upload multiplier to the gas used to store the oracle program.
	afterGas := ctx.GasMeter().GasConsumed()
	gasUsed := afterGas - beforeGas
	var adjGasUsed uint64
	// If the gas used is greater than the max uint64 divided by the upload multiplier we would overflow
	// so we set the gas used to the max uint64, which should result in an out of gas error.
	if gasUsed > math.MaxUint64/params.UploadMultiplier {
		adjGasUsed = math.MaxUint64
	} else {
		adjGasUsed = gasUsed*params.UploadMultiplier - gasUsed
	}
	ctx.GasMeter().ConsumeGas(adjGasUsed, "oracle program upload")

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
