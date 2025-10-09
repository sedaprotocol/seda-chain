package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	vrf "github.com/sedaprotocol/vrf-go"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (m msgServer) LegacyStake(goCtx context.Context, msg *types.MsgLegacyStake) (*types.MsgLegacyStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coreContractAddr, err := m.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}

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

	hash, err := msg.MsgHash(coreContractAddr.String(), ctx.ChainID(), sequenceNum)
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
		return nil, types.ErrInvalidStakerProof.Wrap(err.Error())
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
	return &types.MsgLegacyStakeResponse{}, nil
}

func (m msgServer) LegacyUnstake(goCtx context.Context, msg *types.MsgLegacyUnstake) (*types.MsgLegacyUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coreContractAddr, err := m.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}

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
	hash, err := msg.MsgHash(coreContractAddr.String(), ctx.ChainID(), staker.SequenceNum)
	if err != nil {
		return nil, err
	}
	proof, err := hex.DecodeString(msg.Proof)
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKeyBytes, proof, hash)
	if err != nil {
		return nil, types.ErrInvalidStakerProof.Wrap(err.Error())
	}

	// Update staker info.
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

	return &types.MsgLegacyUnstakeResponse{}, nil
}

func (m msgServer) LegacyWithdraw(goCtx context.Context, msg *types.MsgLegacyWithdraw) (*types.MsgLegacyWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coreContractAddr, err := m.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}

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

	// Verify the proof
	publicKeyBytes, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, err
	}
	hash, err := msg.MsgHash(coreContractAddr.String(), ctx.ChainID(), staker.SequenceNum)
	if err != nil {
		return nil, err
	}
	proof, err := hex.DecodeString(msg.Proof)
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKeyBytes, proof, hash)
	if err != nil {
		return nil, types.ErrInvalidStakerProof.Wrap(err.Error())
	}

	// Update staker info.
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

	return &types.MsgLegacyWithdrawResponse{}, nil
}

func (m msgServer) LegacyCommit(goCtx context.Context, msg *types.MsgLegacyCommit) (*types.MsgLegacyCommitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coreContractAddr, err := m.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}

	paused, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	if paused {
		return nil, types.ErrModulePaused
	}

	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	dr, err := m.GetDataRequest(ctx, msg.DrID)
	if err != nil {
		return nil, err
	}

	// Verify the data request status.
	if dr.Status != types.DATA_REQUEST_STATUS_COMMITTING {
		return nil, types.ErrNotCommitting
	}
	exists, err := m.HasCommitted(ctx, msg.DrID, msg.PublicKey)
	if err != nil {
		return nil, err
	} else if exists {
		return nil, types.ErrAlreadyCommitted
	}
	if dr.TimeoutHeight <= ctx.BlockHeight() {
		return nil, types.ErrCommitTimeout
	}

	// Verify the staker.
	staker, err := m.GetStaker(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if params.StakingConfig.AllowlistEnabled {
		allowlisted, err := m.IsAllowlisted(ctx, msg.PublicKey)
		if err != nil {
			return nil, err
		}
		if !allowlisted {
			return nil, types.ErrNotAllowlisted
		}
	}
	if staker.Staked.LT(params.StakingConfig.MinimumStake) {
		return nil, types.ErrInsufficientStake.Wrapf("%s < %s", staker.Staked, params.StakingConfig.MinimumStake)
	}

	// Verify the proof.
	hash, err := msg.MsgHash(coreContractAddr.String(), ctx.ChainID(), dr.PostedHeight)
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
		return nil, types.ErrInvalidCommitProof.Wrapf("%s", err.Error())
	}

	// Store the commit and start reveal phase if the data request is ready.
	commit, err := hex.DecodeString(msg.Commit)
	if err != nil {
		return nil, err
	}
	commitCount, err := m.AddCommit(ctx, msg.DrID, msg.PublicKey, commit)
	if err != nil {
		return nil, err
	}

	var statusUpdate *types.DataRequestStatus
	if commitCount >= dr.ReplicationFactor {
		newStatus := types.DATA_REQUEST_STATUS_REVEALING
		statusUpdate = &newStatus

		newTimeoutHeight := ctx.BlockHeight() + int64(params.DataRequestConfig.RevealTimeoutInBlocks)
		err = m.UpdateDataRequestTimeout(ctx, msg.DrID, dr.TimeoutHeight, newTimeoutHeight)
		if err != nil {
			return nil, err
		}
		dr.TimeoutHeight = newTimeoutHeight
	}

	err = m.UpdateDataRequest(ctx, &dr, statusUpdate)
	if err != nil {
		return nil, err
	}

	err = m.RefundTxFees(ctx, msg.DrID, msg.PublicKey, false)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommit,
			sdk.NewAttribute(types.AttributeDataRequestID, msg.DrID),
			sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatInt(dr.PostedHeight, 10)),
			sdk.NewAttribute(types.AttributeCommitment, msg.Commit),
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
		),
	)

	return &types.MsgLegacyCommitResponse{}, nil
}

func (m msgServer) LegacyReveal(goCtx context.Context, msg *types.MsgLegacyReveal) (*types.MsgLegacyRevealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coreContractAddr, err := m.wasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}

	paused, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	if paused {
		return nil, types.ErrModulePaused
	}

	// Check the status of the data request.
	dr, err := m.GetDataRequest(ctx, msg.RevealBody.DrID)
	if err != nil {
		return nil, err
	}
	if dr.Status != types.DATA_REQUEST_STATUS_REVEALING {
		return nil, types.ErrRevealNotStarted
	}
	if dr.TimeoutHeight <= ctx.BlockHeight() {
		return nil, types.ErrDataRequestExpired.Wrapf("reveal phase expired at height %d", dr.TimeoutHeight)
	}
	exists, err := m.HasRevealed(ctx, dr.ID, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, types.ErrAlreadyRevealed
	}

	commit, err := m.GetCommit(ctx, dr.ID, msg.PublicKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, types.ErrNotCommitted
		}
		return nil, err
	}

	drConfig, err := m.GetDataRequestConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = msg.Validate(drConfig, dr.ReplicationFactor)
	if err != nil {
		return nil, err
	}

	// Verify against the stored commit.
	expectedCommit, err := msg.RevealHash()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(commit, expectedCommit) {
		return nil, types.ErrRevealMismatch
	}

	// Verify the reveal proof.
	publicKey, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, err
	}
	proof, err := hex.DecodeString(msg.Proof)
	if err != nil {
		return nil, err
	}
	revealHash, err := msg.MsgHash(coreContractAddr.String(), ctx.ChainID())
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKey, proof, revealHash)
	if err != nil {
		return nil, types.ErrInvalidRevealProof.Wrapf("%s", err.Error())
	}

	revealCount, err := m.MarkAsRevealed(ctx, dr.ID, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	var statusUpdate *types.DataRequestStatus
	if revealCount >= dr.ReplicationFactor {
		newStatus := types.DATA_REQUEST_STATUS_TALLYING
		statusUpdate = &newStatus

		err = m.RemoveFromTimeoutQueue(ctx, dr.ID, dr.TimeoutHeight)
		if err != nil {
			return nil, err
		}
	}

	err = m.UpdateDataRequest(ctx, &dr, statusUpdate)
	if err != nil {
		return nil, err
	}

	err = m.SetRevealBody(ctx, msg.PublicKey, *msg.RevealBody)
	if err != nil {
		return nil, err
	}

	err = m.RefundTxFees(ctx, dr.ID, msg.PublicKey, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReveal,
			sdk.NewAttribute(types.AttributeDataRequestID, dr.ID),
			sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatInt(dr.PostedHeight, 10)),
			sdk.NewAttribute(types.AttributeRevealBody, msg.RevealBody.String()),
			sdk.NewAttribute(types.AttributeRevealStdout, strings.Join(msg.Stdout, ",")),
			sdk.NewAttribute(types.AttributeRevealStderr, strings.Join(msg.Stderr, ",")),
			sdk.NewAttribute(types.AttributeExecutorIdentity, msg.PublicKey),
		),
	)

	return &types.MsgLegacyRevealResponse{}, nil
}
