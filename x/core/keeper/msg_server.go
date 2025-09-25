package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	vrf "github.com/sedaprotocol/vrf-go"

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
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", owner, msg.Sender)
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
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
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
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", owner, msg.Sender)
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
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
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
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", owner, msg.Sender)
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
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", owner, msg.Sender)
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

func (m msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if paused
	paused, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	if paused {
		return nil, types.ErrModulePaused
	}

	// Verify stake proof.
	var sequenceNum uint64
	var isExistingStaker bool // for later use
	staker, err := m.GetStaker(ctx, msg.PublicKey)
	if err != nil {
		if !errors.Is(err, collections.ErrNotFound) {
			return nil, err
		}
	} else {
		sequenceNum = staker.SequenceNum
		isExistingStaker = true
	}

	hash, err := msg.MsgHash(ctx.ChainID(), sequenceNum)
	if err != nil {
		return nil, err
	}
	publicKey, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, err
	}
	proof, err := hex.DecodeString(msg.Proof)
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKey, proof, hash)
	if err != nil {
		return nil, types.ErrInvalidStakeProof.Wrapf(err.Error())
	}

	// Verify that the staker is allowlisted if allowlist is enabled.
	stakingConfig, err := m.GetStakingConfig(ctx)
	if err != nil {
		return nil, err
	}
	if stakingConfig.AllowlistEnabled {
		allowlisted, err := m.IsAllowlisted(ctx, msg.PublicKey)
		if err != nil {
			return nil, err
		}
		if !allowlisted {
			return nil, types.ErrNotAllowlisted
		}
	}

	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Stake.Denom != denom {
		return nil, sdkerrors.ErrInvalidCoins.Wrapf("invalid denom: %s", msg.Stake.Denom)
	}

	// Check stake amount and save the staker.
	if isExistingStaker {
		staker.Staked = staker.Staked.Add(msg.Stake.Amount)
		staker.Memo = msg.Memo
	} else {
		if msg.Stake.Amount.LT(stakingConfig.MinimumStake) {
			return nil, types.ErrInsufficientStake.Wrapf("%s < %s", msg.Stake.Amount, stakingConfig.MinimumStake)
		}
		staker = types.Staker{
			PublicKey:         msg.PublicKey,
			Memo:              msg.Memo,
			Staked:            msg.Stake.Amount,
			PendingWithdrawal: math.NewInt(0),
			SequenceNum:       sequenceNum,
		}
	}

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", msg.Sender)
	}
	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(msg.Stake))
	if err != nil {
		return nil, err
	}

	staker.SequenceNum = sequenceNum + 1
	err = m.SetStaker(ctx, staker)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Stake.String()),
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
			sdk.NewAttribute(types.AttributeTokensStaked, staker.Staked.String()),
			sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
			sdk.NewAttribute(types.AttributeMemo, staker.Memo),
			sdk.NewAttribute(types.AttributeSequenceNumber, strconv.FormatUint(sequenceNum, 10)),
		),
	)
	return &types.MsgStakeResponse{}, nil
}

func (m msgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if paused
	paused, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	if paused {
		return nil, types.ErrModulePaused
	}

	// Verify the sender
	_, err = sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", msg.Sender)
	}

	// Check the staker exists
	staker, err := m.GetStaker(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	// Verify the proof
	publicKeyBytes, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, err
	}
	hash, err := msg.MsgHash(ctx.ChainID(), staker.SequenceNum)
	if err != nil {
		return nil, err
	}
	proof, err := hex.DecodeString(msg.Proof)
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKeyBytes, proof, hash)
	if err != nil {
		return nil, types.ErrInvalidStakeProof.Wrapf(err.Error())
	}

	// Update staker info
	sequenceNum := staker.SequenceNum
	unstakeAmount := staker.Staked
	staker.PendingWithdrawal = staker.PendingWithdrawal.Add(unstakeAmount)
	staker.Staked = math.ZeroInt()
	staker.SequenceNum++
	err = m.SetStaker(ctx, staker)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(sdk.AttributeKeyAmount, unstakeAmount.String()),
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
			sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
			sdk.NewAttribute(types.AttributeMemo, staker.Memo),
			sdk.NewAttribute(types.AttributeSequenceNumber, strconv.FormatUint(sequenceNum, 10)),
		),
	)

	return &types.MsgUnstakeResponse{}, nil
}

func (m msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if paused
	paused, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	if paused {
		return nil, types.ErrModulePaused
	}

	// Verify the Sender
	_, err = sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", msg.Sender)
	}

	// Verify the withdraw address
	withdrawAddr, err := sdk.AccAddressFromBech32(msg.WithdrawAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid withdraw address: %s", msg.WithdrawAddress)
	}

	// Check the staker exists
	staker, err := m.GetStaker(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	sequenceNum := staker.SequenceNum
	staker.SequenceNum++

	amount := staker.PendingWithdrawal
	staker.PendingWithdrawal = math.ZeroInt()

	if staker.Staked.IsZero() && staker.PendingWithdrawal.IsZero() {
		err = m.RemoveStaker(ctx, msg.PublicKey)
	} else {
		err = m.SetStaker(ctx, staker)
	}
	if err != nil {
		return nil, err
	}

	// Send coins
	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	err = m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, coins)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdraw,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
			sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
			sdk.NewAttribute(types.AttributeMemo, staker.Memo),
			sdk.NewAttribute(types.AttributeSequenceNumber, strconv.FormatUint(sequenceNum, 10)),
		),
	)

	return &types.MsgWithdrawResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
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
