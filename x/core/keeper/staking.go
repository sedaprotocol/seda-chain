package keeper

import (
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type MsgStake interface {
	Validate() error
	MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte
	GetSender() string
	GetPublicKey() string
	GetMemo() string
	GetProof() string
	GetStake() sdk.Coin
}

func (k Keeper) Stake(ctx sdk.Context, msg MsgStake, isLegacy bool) error {
	err := msg.Validate()
	if err != nil {
		return err
	}

	paused, err := k.IsPaused(ctx)
	if err != nil {
		return err
	}
	if paused {
		return types.ErrModulePaused
	}

	// Verify stake proof.
	var sequenceNum uint64
	var isExistingStaker bool // for later use
	staker, err := k.GetStaker(ctx, msg.GetPublicKey())
	if err != nil {
		if !errors.Is(err, collections.ErrNotFound) {
			return err
		}
	} else {
		sequenceNum = staker.SequenceNum
		isExistingStaker = true
	}

	var coreContractAddr string
	if isLegacy {
		addr, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
		if err != nil {
			return err
		}
		coreContractAddr = addr.String()
	}

	err = VerifyStakerProof(msg.GetPublicKey(), msg.GetProof(), msg.MsgHash(coreContractAddr, ctx.ChainID(), sequenceNum))
	if err != nil {
		return types.ErrInvalidStakerProof.Wrap(err.Error())
	}

	// Verify that the staker is allowlisted if allowlist is enabled.
	stakingConfig, err := k.GetStakingConfig(ctx)
	if err != nil {
		return err
	}
	if stakingConfig.AllowlistEnabled {
		allowlisted, err := k.IsAllowlisted(ctx, msg.GetPublicKey())
		if err != nil {
			return err
		}
		if !allowlisted {
			return types.ErrNotAllowlisted
		}
	}

	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	stake := msg.GetStake()
	if stake.Denom != denom {
		return sdkerrors.ErrInvalidCoins.Wrapf("invalid denom: %s", stake.Denom)
	}

	// Check stake amount and save the staker.
	if isExistingStaker {
		staker.Staked = staker.Staked.Add(stake.Amount)
		staker.Memo = msg.GetMemo()
	} else {
		if stake.Amount.LT(stakingConfig.MinimumStake) {
			return types.ErrInsufficientStake.Wrapf("%s < %s", stake.Amount, stakingConfig.MinimumStake)
		}
		staker = types.Staker{
			PublicKey:         msg.GetPublicKey(),
			Memo:              msg.GetMemo(),
			Staked:            stake.Amount,
			PendingWithdrawal: math.NewInt(0),
			SequenceNum:       sequenceNum,
		}
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.GetSender()), types.ModuleName, sdk.NewCoins(stake))
	if err != nil {
		return err
	}

	staker.SequenceNum = sequenceNum + 1
	err = k.SetStaker(ctx, staker)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorActionEvent,
			sdk.NewAttribute(types.AttributeAction, types.EventTypeStake),
			sdk.NewAttribute(types.AttributeExecutorIdentity, staker.PublicKey),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSender()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, stake.Amount.String()),
			sdk.NewAttribute(types.AttributeSequenceNumber, strconv.FormatUint(sequenceNum, 10)),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorEvent,
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.GetPublicKey()),
			sdk.NewAttribute(types.AttributeTokensStaked, staker.Staked.String()),
			sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
			sdk.NewAttribute(types.AttributeMemo, staker.Memo),
		),
	)
	return nil
}

type MsgUnstake interface {
	GetSender() string
	GetPublicKey() string
	GetProof() string
	MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte
}

func (k Keeper) Unstake(ctx sdk.Context, msg MsgUnstake, isLegacy bool) error {
	paused, err := k.IsPaused(ctx)
	if err != nil {
		return err
	}
	if paused {
		return types.ErrModulePaused
	}

	// Verify the sender
	_, err = sdk.AccAddressFromBech32(msg.GetSender())
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", msg.GetSender())
	}

	// Verify the staker and update its info.
	staker, err := k.GetStaker(ctx, msg.GetPublicKey())
	if err != nil {
		return err
	}

	var coreContractAddr string
	if isLegacy {
		addr, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
		if err != nil {
			return err
		}
		coreContractAddr = addr.String()
	}

	err = VerifyStakerProof(msg.GetPublicKey(), msg.GetProof(), msg.MsgHash(coreContractAddr, ctx.ChainID(), staker.SequenceNum))
	if err != nil {
		return types.ErrInvalidStakerProof.Wrap(err.Error())
	}

	unstakeAmount := staker.Staked
	staker.PendingWithdrawal = staker.PendingWithdrawal.Add(unstakeAmount)
	staker.Staked = math.ZeroInt()

	sequenceNum := staker.SequenceNum
	staker.SequenceNum++

	err = k.SetStaker(ctx, staker)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorActionEvent,
			sdk.NewAttribute(types.AttributeAction, types.EventTypeUnstake),
			sdk.NewAttribute(types.AttributeExecutorIdentity, staker.PublicKey),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSender()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, unstakeAmount.String()),
			sdk.NewAttribute(types.AttributeSequenceNumber, strconv.FormatUint(sequenceNum, 10)),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorEvent,
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.GetPublicKey()),
			sdk.NewAttribute(types.AttributeTokensStaked, staker.Staked.String()),
			sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
			sdk.NewAttribute(types.AttributeMemo, staker.Memo),
		),
	)

	return nil
}

type MsgWithdraw interface {
	GetSender() string
	GetPublicKey() string
	GetProof() string
	GetWithdrawAddress() string
	MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte
}

func (k Keeper) Withdraw(ctx sdk.Context, msg MsgWithdraw, isLegacy bool) error {
	paused, err := k.IsPaused(ctx)
	if err != nil {
		return err
	}
	if paused {
		return types.ErrModulePaused
	}

	// Verify the Sender
	_, err = sdk.AccAddressFromBech32(msg.GetSender())
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", msg.GetSender())
	}

	// Verify the withdraw address
	withdrawAddr, err := sdk.AccAddressFromBech32(msg.GetWithdrawAddress())
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid withdraw address: %s", msg.GetWithdrawAddress())
	}

	// Verify the staker and update its info.
	staker, err := k.GetStaker(ctx, msg.GetPublicKey())
	if err != nil {
		return err
	}

	var coreContractAddr string
	if isLegacy {
		addr, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
		if err != nil {
			return err
		}
		coreContractAddr = addr.String()
	}

	err = VerifyStakerProof(msg.GetPublicKey(), msg.GetProof(), msg.MsgHash(coreContractAddr, ctx.ChainID(), staker.SequenceNum))
	if err != nil {
		return types.ErrInvalidStakerProof.Wrap(err.Error())
	}

	amount := staker.PendingWithdrawal
	staker.PendingWithdrawal = math.ZeroInt()

	sequenceNum := staker.SequenceNum
	staker.SequenceNum++

	if staker.Staked.IsZero() && staker.PendingWithdrawal.IsZero() {
		err = k.RemoveStaker(ctx, msg.GetPublicKey())
	} else {
		err = k.SetStaker(ctx, staker)
	}
	if err != nil {
		return err
	}

	// Send coins
	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}
	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, coins)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorActionEvent,
			sdk.NewAttribute(types.AttributeAction, types.EventTypeWithdraw),
			sdk.NewAttribute(types.AttributeExecutorIdentity, staker.PublicKey),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetSender()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeSequenceNumber, strconv.FormatUint(sequenceNum, 10)),
		),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecutorEvent,
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.GetPublicKey()),
			sdk.NewAttribute(types.AttributeTokensStaked, staker.Staked.String()),
			sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
			sdk.NewAttribute(types.AttributeMemo, staker.Memo),
		),
	)

	return nil
}
