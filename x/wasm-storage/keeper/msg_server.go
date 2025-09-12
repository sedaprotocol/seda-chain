package keeper

import (
	"context"
	"encoding/hex"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

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

	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	// The user should have attached fees for the size of the wasm file multiplied
	// by the cost per byte, so we derive the max wasm size from that.
	storageFee := msg.StorageFee.AmountOf(denom)
	paidStorage := storageFee.Quo(math.NewIntFromUint64(params.WasmCostPerByte))

	if !paidStorage.IsInt64() || paidStorage.Int64() > params.MaxWasmSize {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "WASM file is too large")
	}

	senderAddress, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddress, authtypes.FeeCollectorName, msg.StorageFee)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	unzipped, err := unzipWasm(msg.Wasm, paidStorage.Int64())
	if err != nil {
		return nil, err
	}

	program := types.NewOracleProgram(unzipped, ctx.BlockTime())
	if exists, _ := m.OracleProgram.Has(ctx, program.Hash); exists {
		return nil, types.ErrWasmAlreadyExists
	}

	if err := m.OracleProgram.Set(ctx, program.Hash, program); err != nil {
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

	return &types.MsgStoreOracleProgramResponse{
		Hash: hashHex,
	}, nil
}

// InstantiateCoreContract instantiates a new contract with a
// predictable address and updates the core contract registry.
func (m msgServer) InstantiateCoreContract(goCtx context.Context, msg *types.MsgInstantiateCoreContract) (*types.MsgInstantiateCoreContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if m.GetAuthority() != msg.Sender {
		return nil, types.ErrInvalidAuthority.Wrapf("expected %s, got %s", m.authority, msg.Sender)
	}

	var adminAddr sdk.AccAddress
	var err error
	if msg.Admin != "" {
		if adminAddr, err = sdk.AccAddressFromBech32(msg.Admin); err != nil {
			return nil, err
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
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

// UpdateParams updates the module parameters.
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
