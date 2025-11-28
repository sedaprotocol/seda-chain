package keeper

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type MsgCommit interface {
	GetDrID() string
	GetCommit() string
	GetPublicKey() string
	GetProof() string
	MsgHash(coreContractAddr, chainID string, drHeight int64) []byte
}

func (k Keeper) Commit(ctx sdk.Context, msg MsgCommit, isLegacy bool) error {
	drID := msg.GetDrID()
	publicKey := msg.GetPublicKey()

	paused, err := k.IsPaused(ctx)
	if err != nil {
		return err
	}
	if paused {
		return types.ErrModulePaused
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	dr, err := k.GetDataRequest(ctx, drID)
	if err != nil {
		return err
	}

	// Verify the data request status.
	if dr.Status != types.DATA_REQUEST_STATUS_COMMITTING {
		return types.ErrNotCommitting
	}
	exists, err := k.HasCommitted(ctx, drID, publicKey)
	if err != nil {
		return err
	} else if exists {
		return types.ErrAlreadyCommitted
	}
	if dr.TimeoutHeight <= ctx.BlockHeight() {
		return types.ErrCommitTimeout
	}

	// Verify the staker and update its info.
	staker, err := k.GetStaker(ctx, publicKey)
	if err != nil {
		return err
	}

	if params.StakingConfig.AllowlistEnabled {
		allowlisted, err := k.IsAllowlisted(ctx, publicKey)
		if err != nil {
			return err
		}
		if !allowlisted {
			return types.ErrNotAllowlisted
		}
	}
	if staker.Staked.LT(params.StakingConfig.MinimumStake) {
		return types.ErrInsufficientStake.Wrapf("%s < %s", staker.Staked, params.StakingConfig.MinimumStake)
	}

	var coreContractAddr string
	if isLegacy {
		addr, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
		if err != nil {
			return err
		}
		coreContractAddr = addr.String()
	}

	err = VerifyStakerProof(publicKey, msg.GetProof(), msg.MsgHash(coreContractAddr, ctx.ChainID(), dr.PostedHeight))
	if err != nil {
		return types.ErrInvalidCommitProof.Wrap(err.Error())
	}

	// Store the commit and start reveal phase if the data request is ready.
	commit, err := hex.DecodeString(msg.GetCommit())
	if err != nil {
		return err
	}
	commitCount, err := k.AddCommit(ctx, drID, publicKey, commit)
	if err != nil {
		return err
	}

	var statusUpdate *types.DataRequestStatus
	if commitCount >= dr.ReplicationFactor {
		newStatus := types.DATA_REQUEST_STATUS_REVEALING
		statusUpdate = &newStatus

		newTimeoutHeight := ctx.BlockHeight() + int64(params.DataRequestConfig.RevealTimeoutInBlocks)
		err = k.UpdateDataRequestTimeout(ctx, drID, dr.TimeoutHeight, newTimeoutHeight)
		if err != nil {
			return err
		}
		dr.TimeoutHeight = newTimeoutHeight
	}

	err = k.UpdateDataRequest(ctx, &dr, statusUpdate)
	if err != nil {
		return err
	}

	err = k.RefundTxFees(ctx, drID, publicKey, false)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCommit,
			sdk.NewAttribute(types.AttributeDataRequestID, drID),
			sdk.NewAttribute(types.AttributePostedDataRequestHeight, strconv.FormatInt(dr.PostedHeight, 10)),
			sdk.NewAttribute(types.AttributeCommitment, msg.GetCommit()),
			sdk.NewAttribute(types.AttributeExecutor, publicKey),
		),
	)

	return nil
}

type MsgReveal interface {
	GetRevealBody() *types.RevealBody
	GetPublicKey() string
	GetProof() string
	GetStderr() []string
	GetStdout() []string
	MsgHash(coreContractAddr, chainID string) []byte
	RevealHash() []byte
	Validate(config types.DataRequestConfig, replicationFactor uint32) error
}

func (k Keeper) Reveal(ctx sdk.Context, msg MsgReveal, isLegacy bool) error {
	paused, err := k.IsPaused(ctx)
	if err != nil {
		return err
	}
	if paused {
		return types.ErrModulePaused
	}

	// Check the status of the data request.
	revealBody := msg.GetRevealBody()

	dr, err := k.GetDataRequest(ctx, revealBody.DrID)
	if err != nil {
		return err
	}
	if dr.Status != types.DATA_REQUEST_STATUS_REVEALING {
		return types.ErrRevealNotStarted
	}
	if dr.TimeoutHeight <= ctx.BlockHeight() {
		return types.ErrDataRequestExpired.Wrapf("reveal phase expired at height %d", dr.TimeoutHeight)
	}
	publicKey := msg.GetPublicKey()
	exists, err := k.HasRevealed(ctx, dr.ID, publicKey)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrAlreadyRevealed
	}

	// Verify against the stored commit.
	commit, err := k.GetCommit(ctx, dr.ID, publicKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrNotCommitted
		}
		return err
	}
	if !bytes.Equal(commit, msg.RevealHash()) {
		return types.ErrRevealMismatch
	}

	// Validate the reveal body.
	drConfig, err := k.GetDataRequestConfig(ctx)
	if err != nil {
		return err
	}
	err = msg.Validate(drConfig, dr.ReplicationFactor)
	if err != nil {
		return err
	}

	// Verify the reveal proof.
	var coreContractAddr string
	if isLegacy {
		addr, err := k.wasmStorageKeeper.GetCoreContractAddr(ctx)
		if err != nil {
			return err
		}
		coreContractAddr = addr.String()
	}

	err = VerifyStakerProof(publicKey, msg.GetProof(), msg.MsgHash(coreContractAddr, ctx.ChainID()))
	if err != nil {
		return types.ErrInvalidRevealProof.Wrap(err.Error())
	}

	// Store the reveal.
	revealCount, err := k.MarkAsRevealed(ctx, dr.ID, publicKey)
	if err != nil {
		return err
	}

	var statusUpdate *types.DataRequestStatus
	if revealCount >= dr.ReplicationFactor {
		newStatus := types.DATA_REQUEST_STATUS_TALLYING
		statusUpdate = &newStatus

		err = k.RemoveFromTimeoutQueue(ctx, dr.ID, dr.TimeoutHeight)
		if err != nil {
			return err
		}
	}

	err = k.UpdateDataRequest(ctx, &dr, statusUpdate)
	if err != nil {
		return err
	}

	err = k.SetRevealBody(ctx, publicKey, *revealBody)
	if err != nil {
		return err
	}

	err = k.RefundTxFees(ctx, dr.ID, publicKey, true)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReveal,
			sdk.NewAttribute(types.AttributeDataRequestID, dr.ID),
			sdk.NewAttribute(types.AttributePostedDataRequestHeight, strconv.FormatInt(dr.PostedHeight, 10)),
			sdk.NewAttribute(types.AttributeRevealExitCode, strconv.FormatUint(uint64(revealBody.ExitCode), 10)),
			sdk.NewAttribute(types.AttributeRevealGasUsed, strconv.FormatUint(revealBody.GasUsed, 10)),
			sdk.NewAttribute(types.AttributeReveal, base64.StdEncoding.EncodeToString(revealBody.Reveal)),
			sdk.NewAttribute(types.AttributeRevealProxyPubKeys, strings.Join(revealBody.ProxyPubKeys, ",")),
			sdk.NewAttribute(types.AttributeRevealStdout, strings.Join(msg.GetStdout(), ",")),
			sdk.NewAttribute(types.AttributeRevealStderr, strings.Join(msg.GetStderr(), ",")),
			sdk.NewAttribute(types.AttributeExecutor, publicKey),
		),
	)

	return nil
}
