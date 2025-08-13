package keeper

import (
	"bytes"
	"context"
	"encoding/hex"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
	vrf "github.com/sedaprotocol/vrf-go"
)

func (m msgServer) PostDataRequest(goCtx context.Context, msg *types.MsgPostDataRequest) (*types.MsgPostDataRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	drConfig, err := m.GetDataRequestConfig(ctx)
	if err != nil {
		return nil, err
	}
	if err := msg.Validate(drConfig); err != nil {
		return nil, err
	}

	count, err := m.GetStakersCount(ctx)
	if err != nil {
		return nil, err
	}
	maxRF := min(count, types.MaxReplicationFactor)
	if msg.ReplicationFactor > uint32(maxRF) {
		return nil, types.ErrReplicationFactorTooHigh.Wrapf("%d > %d", msg.ReplicationFactor, maxRF)
	}

	drID, err := msg.MsgHash()
	if err != nil {
		return nil, err
	}
	exists, err := m.HasDataRequest(ctx, drID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, types.ErrDataRequestAlreadyExists
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

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(msg.Funds))
	if err != nil {
		return nil, err
	}

	dr := types.DataRequest{
		Id:                drID,
		Version:           msg.Version,
		ExecProgramId:     msg.ExecProgramId,
		ExecInputs:        msg.ExecInputs,
		ExecGasLimit:      msg.ExecGasLimit,
		TallyProgramId:    msg.TallyProgramId,
		TallyInputs:       msg.TallyInputs,
		TallyGasLimit:     msg.TallyGasLimit,
		ReplicationFactor: msg.ReplicationFactor,
		ConsensusFilter:   msg.ConsensusFilter,
		GasPrice:          msg.GasPrice,
		Memo:              msg.Memo,
		PaybackAddress:    msg.PaybackAddress,
		SedaPayload:       msg.SedaPayload,
		Height:            uint64(ctx.BlockHeight()),
		PostedGasPrice:    msg.GasPrice,
		Poster:            msg.Sender,
		Escrow:            msg.Funds.Amount,
		TimeoutHeight:     uint64(ctx.BlockHeight()) + uint64(drConfig.CommitTimeoutInBlocks),
		Status:            types.DATA_REQUEST_COMMITTING,
		// Commits:           make(map[string][]byte), // Dropped by proto anyways
		// Reveals:           make(map[string]bool), // Dropped by proto anyways
	}
	err = m.SetDataRequest(ctx, dr)
	if err != nil {
		return nil, err
	}

	err = m.AddToCommitting(ctx, dr.Index())
	if err != nil {
		return nil, err
	}

	err = m.timeoutQueue.Set(ctx, collections.Join(dr.TimeoutHeight, drID))
	if err != nil {
		return nil, err
	}

	// TODO emit events

	return &types.MsgPostDataRequestResponse{
		DrId:   drID,
		Height: dr.Height,
	}, nil
}

func (m msgServer) Commit(goCtx context.Context, msg *types.MsgCommit) (*types.MsgCommitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := m.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	dr, err := m.GetDataRequest(ctx, msg.DrId)
	if err != nil {
		return nil, err
	}

	// Verify the data request status.
	if dr.Status != types.DATA_REQUEST_COMMITTING {
		return nil, types.ErrNotCommitting
	}
	if _, ok := dr.Commits[msg.PublicKey]; ok {
		return nil, types.ErrAlreadyCommitted
	}
	if dr.TimeoutHeight <= uint64(ctx.BlockHeight()) {
		return nil, types.ErrCommitTimeout
	}

	// Verify the staker.
	staker, err := m.Stakers.Get(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if staker.Staked.LT(params.StakingConfig.MinimumStake) {
		return nil, types.ErrInsufficientStake.Wrapf("%s < %s", staker.Staked, params.StakingConfig.MinimumStake)
	}

	// Verify the proof.
	hash, err := msg.MsgHash("", ctx.ChainID(), dr.Height)
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
		return nil, types.ErrInvalidCommitProof.Wrapf(err.Error())
	}

	// Add the commitment and start reveal phase if the data request is ready.
	commitment, err := hex.DecodeString(msg.Commitment)
	if err != nil {
		return nil, err
	}
	dr.AddCommit(msg.PublicKey, commitment)

	if len(dr.Commits) >= int(dr.ReplicationFactor) {
		dr.Status = types.DATA_REQUEST_REVEALING

		newTimeoutHeight := dr.TimeoutHeight + uint64(params.DataRequestConfig.RevealTimeoutInBlocks)
		err = m.UpdateDataRequestTimeout(ctx, msg.DrId, dr.TimeoutHeight, newTimeoutHeight)
		if err != nil {
			return nil, err
		}
		dr.TimeoutHeight = newTimeoutHeight

		err = m.CommittingToRevealing(ctx, dr.Index())
		if err != nil {
			return nil, err
		}
	}

	err = m.SetDataRequest(ctx, dr)
	if err != nil {
		return nil, err
	}

	// TODO Refund (ref https://github.com/sedaprotocol/seda-chain/pull/527)
	// TODO Emit events

	return &types.MsgCommitResponse{}, nil
}

func (m msgServer) Reveal(goCtx context.Context, msg *types.MsgReveal) (*types.MsgRevealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check the status of the data request.
	dr, err := m.GetDataRequest(ctx, msg.RevealBody.DrId)
	if err != nil {
		return nil, err
	}
	if dr.Status != types.DATA_REQUEST_REVEALING {
		return nil, types.ErrRevealNotStarted
	}
	if dr.TimeoutHeight <= uint64(ctx.BlockHeight()) {
		return nil, types.ErrDataRequestExpired.Wrapf("reveal phase expired at height %d", dr.TimeoutHeight)
	}
	if dr.HasRevealed(msg.PublicKey) {
		return nil, types.ErrAlreadyRevealed
	}

	commit, exists := dr.GetCommit(msg.PublicKey)
	if !exists {
		return nil, types.ErrNotCommitted
	}

	// Check reveal size limit.
	drConfig, err := m.GetDataRequestConfig(ctx)
	if err != nil {
		return nil, err
	}
	revealSizeLimit := drConfig.DrRevealSizeLimitInBytes / dr.ReplicationFactor
	if len(msg.RevealBody.Reveal) > int(revealSizeLimit) {
		return nil, types.ErrRevealTooBig.Wrapf("%d bytes > %d bytes", len(msg.RevealBody.Reveal), revealSizeLimit)
	}

	// Verify against the stored commit.
	expectedCommit, err := msg.RevealHash()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(commit, expectedCommit) {
		return nil, types.ErrRevealMismatch
	}

	// TODO move to msg.Validate()
	for _, key := range msg.RevealBody.ProxyPubKeys {
		_, err := hex.DecodeString(key)
		if err != nil {
			return nil, err
		}
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
	revealHash, err := msg.MsgHash("", ctx.ChainID())
	if err != nil {
		return nil, err
	}
	_, err = vrf.NewK256VRF().Verify(publicKey, proof, revealHash)
	if err != nil {
		return nil, types.ErrInvalidRevealProof.Wrapf(err.Error())
	}

	revealsCount := dr.MarkAsRevealed(msg.PublicKey)
	if revealsCount >= int(dr.ReplicationFactor) {
		dr.Status = types.DATA_REQUEST_TALLYING

		err = m.RemoveFromTimeoutQueue(ctx, dr.Id, dr.TimeoutHeight)
		if err != nil {
			return nil, err
		}

		err = m.RevealingToTallying(ctx, dr.Index())
		if err != nil {
			return nil, err
		}
	}

	err = m.SetRevealBody(ctx, dr.Id, msg.PublicKey, *msg.RevealBody)
	if err != nil {
		return nil, err
	}
	err = m.SetDataRequest(ctx, dr)
	if err != nil {
		return nil, err
	}

	// TODO: Add refund logic
	// TODO Emit events

	return &types.MsgRevealResponse{}, nil
}
