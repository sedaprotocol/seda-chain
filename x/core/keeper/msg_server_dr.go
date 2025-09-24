package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	vrf "github.com/sedaprotocol/vrf-go"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (m msgServer) PostDataRequest(goCtx context.Context, msg *types.MsgPostDataRequest) (*types.MsgPostDataRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if paused
	paused, err := m.IsPaused(ctx)
	if err != nil {
		return nil, err
	}
	if paused {
		return nil, types.ErrModulePaused
	}

	drConfig, err := m.GetDataRequestConfig(ctx)
	if err != nil {
		return nil, err
	}
	err = msg.Validate(drConfig)
	if err != nil {
		return nil, err
	}

	count, err := m.GetStakerCount(ctx)
	if err != nil {
		return nil, err
	}
	maxRF := min(count, types.MaxReplicationFactor)
	if msg.ReplicationFactor > maxRF {
		return nil, types.ErrReplicationFactorTooHigh.Wrapf("%d > %d", msg.ReplicationFactor, maxRF)
	}

	drID, err := msg.MsgHash()
	if err != nil {
		return nil, err
	}

	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	if msg.Funds.Denom != denom {
		return nil, sdkerrors.ErrInvalidCoins.Wrapf("invalid denom: %s", msg.Funds.Denom)
	}

	totalGasLimit := math.NewIntFromUint64(msg.ExecGasLimit).Add(math.NewIntFromUint64(msg.TallyGasLimit))
	postedGasPrice := msg.Funds.Amount.Quo(totalGasLimit)
	if postedGasPrice.LT(msg.GasPrice) {
		requiredFunds, _ := totalGasLimit.SafeMul(msg.GasPrice)
		return nil, sdkerrors.ErrInsufficientFunds.Wrapf("required: %s, got %s", requiredFunds, msg.GasPrice)
	}

	err = m.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sdk.MustAccAddressFromBech32(msg.Sender), // already validated in msg.Validate()
		types.ModuleName,
		sdk.NewCoins(msg.Funds),
	)
	if err != nil {
		return nil, err
	}

	dr := types.DataRequest{
		ID:                drID,
		Version:           msg.Version,
		ExecProgramID:     msg.ExecProgramID,
		ExecInputs:        msg.ExecInputs,
		ExecGasLimit:      msg.ExecGasLimit,
		TallyProgramID:    msg.TallyProgramID,
		TallyInputs:       msg.TallyInputs,
		TallyGasLimit:     msg.TallyGasLimit,
		ReplicationFactor: msg.ReplicationFactor,
		ConsensusFilter:   msg.ConsensusFilter,
		GasPrice:          msg.GasPrice,
		Memo:              msg.Memo,
		PaybackAddress:    msg.PaybackAddress,
		SEDAPayload:       msg.SEDAPayload,
		PostedHeight:      ctx.BlockHeight(),
		PostedGasPrice:    postedGasPrice,
		Poster:            msg.Sender,
		Escrow:            msg.Funds.Amount,
		Status:            types.DATA_REQUEST_STATUS_COMMITTING,
		TimeoutHeight:     ctx.BlockHeight() + int64(drConfig.CommitTimeoutInBlocks),
		// Commits:           make(map[string][]byte), // Dropped by proto anyways
		// Reveals:           make(map[string]bool), // Dropped by proto anyways
	}

	err = m.AddToTimeoutQueue(ctx, drID, dr.TimeoutHeight)
	if err != nil {
		return nil, err
	}

	err = m.StoreDataRequest(ctx, dr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePostDataRequest,
			sdk.NewAttribute(types.AttributeDataRequestID, dr.ID),
			sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatInt(dr.PostedHeight, 10)),
			sdk.NewAttribute(types.AttributeExecutorIdentity, dr.Poster),
			sdk.NewAttribute(types.AttributeExecProgramID, dr.ExecProgramID),
			sdk.NewAttribute(types.AttributeExecInputs, base64.StdEncoding.EncodeToString(dr.ExecInputs)),
			sdk.NewAttribute(types.AttributeExecGasLimit, strconv.FormatUint(dr.ExecGasLimit, 10)),
			sdk.NewAttribute(types.AttributeTallyProgramID, dr.TallyProgramID),
			sdk.NewAttribute(types.AttributeTallyInputs, base64.StdEncoding.EncodeToString(dr.TallyInputs)),
			sdk.NewAttribute(types.AttributeTallyGasLimit, strconv.FormatUint(dr.TallyGasLimit, 10)),
			sdk.NewAttribute(types.AttributeReplicationFactor, strconv.FormatUint(uint64(dr.ReplicationFactor), 10)),
			sdk.NewAttribute(types.AttributeConsensusFilter, base64.StdEncoding.EncodeToString(dr.ConsensusFilter)),
			sdk.NewAttribute(types.AttributeGasPrice, dr.GasPrice.String()),
			sdk.NewAttribute(types.AttributeMemo, base64.StdEncoding.EncodeToString(dr.Memo)),
			sdk.NewAttribute(types.AttributeSEDAPayload, base64.StdEncoding.EncodeToString(dr.SEDAPayload)),
			sdk.NewAttribute(types.AttributePaybackAddress, base64.StdEncoding.EncodeToString(dr.PaybackAddress)),
			sdk.NewAttribute(types.AttributeVersion, dr.Version),
			sdk.NewAttribute(types.AttributePostedGasPrice, dr.PostedGasPrice.String()),
		),
	)

	return &types.MsgPostDataRequestResponse{
		DrID:   drID,
		Height: dr.PostedHeight,
	}, nil
}

func (m msgServer) Commit(goCtx context.Context, msg *types.MsgCommit) (*types.MsgCommitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if paused
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
	if _, ok := dr.Commits[msg.PublicKey]; ok {
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
	hash, err := msg.MsgHash(ctx.ChainID(), dr.PostedHeight)
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
	dr.AddCommit(msg.PublicKey, commit)

	var statusUpdate *types.DataRequestStatus
	if len(dr.Commits) >= int(dr.ReplicationFactor) {
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

	return &types.MsgCommitResponse{}, nil
}

func (m msgServer) Reveal(goCtx context.Context, msg *types.MsgReveal) (*types.MsgRevealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if paused
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
	if dr.HasRevealed(msg.PublicKey) {
		return nil, types.ErrAlreadyRevealed
	}

	commit, exists := dr.GetCommit(msg.PublicKey)
	if !exists {
		return nil, types.ErrNotCommitted
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
	revealHash, err := msg.MsgHash(ctx.ChainID())
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKey, proof, revealHash)
	if err != nil {
		return nil, types.ErrInvalidRevealProof.Wrapf("%s", err.Error())
	}

	revealsCount := dr.MarkAsRevealed(msg.PublicKey)

	var statusUpdate *types.DataRequestStatus
	if revealsCount >= int(dr.ReplicationFactor) {
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

	return &types.MsgRevealResponse{}, nil
}
