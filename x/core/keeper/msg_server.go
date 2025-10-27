package keeper

import (
	"context"
	"encoding/base64"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (m msgServer) PostDataRequest(goCtx context.Context, msg *types.MsgPostDataRequest) (*types.MsgPostDataRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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

	drID, err := msg.ComputeDataRequestID()
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
			sdk.NewAttribute(types.AttributeDataRequestPoster, dr.Poster),
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
