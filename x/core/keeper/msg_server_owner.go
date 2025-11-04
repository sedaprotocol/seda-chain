package keeper

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AcceptOwnership(goCtx context.Context, msg *types.MsgAcceptOwnership) (*types.MsgAcceptOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	currentPendingOwner, err := m.GetPendingOwner(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Sender != currentPendingOwner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", currentPendingOwner, msg.Sender)
	}

	err = m.SetOwner(ctx, msg.Sender)
	if err != nil {
		return nil, err
	}

	err = m.SetPendingOwner(ctx, "")
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAcceptOwnership,
			sdk.NewAttribute(types.AttributeNewOwner, msg.Sender),
		),
	)

	return &types.MsgAcceptOwnershipResponse{}, nil
}

func (m msgServer) TransferOwnership(goCtx context.Context, msg *types.MsgTransferOwnership) (*types.MsgTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	currentOwner, err := m.GetOwner(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Sender != currentOwner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", currentOwner, msg.Sender)
	}

	// validate new owner address
	if _, err := sdk.AccAddressFromBech32(msg.NewOwner); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", msg.NewOwner)
	}

	err = m.SetPendingOwner(ctx, msg.NewOwner)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferOwnership,
			sdk.NewAttribute(types.AttributePendingOwner, msg.NewOwner),
		),
	)

	return &types.MsgTransferOwnershipResponse{}, nil
}

func (m msgServer) AddToAllowlist(goCtx context.Context, msg *types.MsgAddToAllowlist) (*types.MsgAddToAllowlistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := m.GetOwner(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Sender != owner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", owner, msg.Sender)
	}

	// TODO: validate public key format
	if msg.PublicKey == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("public key is empty")
	}

	exists, err := m.IsAllowlisted(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, types.ErrAlreadyAllowlisted
	}

	err = m.Keeper.AddToAllowlist(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddToAllowlist,
			sdk.NewAttribute(types.AttributeExecutor, msg.PublicKey),
		),
	)

	return &types.MsgAddToAllowlistResponse{}, nil
}

func (m msgServer) RemoveFromAllowlist(goCtx context.Context, msg *types.MsgRemoveFromAllowlist) (*types.MsgRemoveFromAllowlistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := m.GetOwner(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Sender != owner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", owner, msg.Sender)
	}

	if msg.PublicKey == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("public key is empty")
	}

	exists, err := m.IsAllowlisted(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, types.ErrNotAllowlisted
	}

	err = m.Keeper.RemoveFromAllowlist(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	if staker, err := m.GetStaker(ctx, msg.PublicKey); err == nil {
		if staker.Staked.GT(math.ZeroInt()) {
			staker.PendingWithdrawal = staker.PendingWithdrawal.Add(staker.Staked)
			staker.Staked = math.ZeroInt()
			err = m.SetStaker(ctx, staker)
			if err != nil {
				return nil, err
			}
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveFromAllowlist,
			sdk.NewAttribute(types.AttributeExecutor, msg.PublicKey),
		),
	)

	return &types.MsgRemoveFromAllowlistResponse{}, nil
}

func (m msgServer) Pause(goCtx context.Context, msg *types.MsgPause) (*types.MsgPauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := m.GetOwner(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Sender != owner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", owner, msg.Sender)
	}

	current, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}

	if current {
		return nil, types.ErrModuleAlreadyPaused
	}

	err = m.Keeper.Pause(ctx)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePause,
			sdk.NewAttribute(types.AttributePaused, "true"),
		),
	)

	return &types.MsgPauseResponse{}, nil
}

func (m msgServer) Unpause(goCtx context.Context, msg *types.MsgUnpause) (*types.MsgUnpauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := m.GetOwner(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Sender != owner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", owner, msg.Sender)
	}

	current, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}

	if !current {
		return nil, types.ErrModuleAlreadyUnpaused
	}

	err = m.Keeper.Unpause(ctx)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnpause,
			sdk.NewAttribute(types.AttributePaused, "false"),
		),
	)

	return &types.MsgUnpauseResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := m.GetOwner(ctx)
	if err != nil {
		return nil, err
	}
	if owner != msg.Owner {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized owner; expected %s, got %s", owner, msg.Owner)
	}

	if msg.Params.DataRequestConfig == nil {
		currentDataRequestConfig, err := m.GetDataRequestConfig(ctx)
		if err != nil {
			return nil, err
		}
		msg.Params.DataRequestConfig = &currentDataRequestConfig
	}

	if msg.Params.StakingConfig == nil {
		currentStakingConfig, err := m.GetStakingConfig(ctx)
		if err != nil {
			return nil, err
		}
		msg.Params.StakingConfig = &currentStakingConfig
	}

	if msg.Params.TallyConfig == nil {
		currentTallyConfig, err := m.GetTallyConfig(ctx)
		if err != nil {
			return nil, err
		}
		msg.Params.TallyConfig = &currentTallyConfig
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
