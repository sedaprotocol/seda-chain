package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (m msgServer) PostDataRequest(goCtx context.Context, msg *types.MsgPostDataRequest) (*types.MsgPostDataRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := m.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := msg.Validate(params.DataRequestConfig); err != nil {
		return nil, err
	}

	// TODO Separately store the stakers count?
	count, err := m.GetStakersCount(ctx)
	if err != nil {
		return nil, err
	}
	maxRF := min(count, types.MaxReplicationFactor)
	if msg.ReplicationFactor > uint32(maxRF) {
		return nil, types.ErrReplicationFactorTooHigh.Wrapf("%d > %d", msg.ReplicationFactor, maxRF)
	}

	drID, err := msg.TryHash()
	if err != nil {
		return nil, err
	}
	exists, err := m.DataRequests.Has(ctx, drID)
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
		Commits:           make(map[string][]byte),
		Reveals:           make(map[string]bool),
		Poster:            msg.Sender,
		Escrow:            msg.Funds,
		TimeoutHeight:     uint64(ctx.BlockHeight()) + uint64(params.DataRequestConfig.CommitTimeoutInBlocks),
	}
	err = m.DataRequests.Set(ctx, drID, dr)
	if err != nil {
		return nil, err
	}

	err = m.AddToCommitting(ctx, NewDataRequestIndex(drID, dr.PostedGasPrice, dr.Height))
	if err != nil {
		return nil, err
	}

	err = m.timeoutQueue.Set(ctx, int64(dr.TimeoutHeight), drID)
	if err != nil {
		return nil, err
	}

	// TODO emit events

	return &types.MsgPostDataRequestResponse{
		DrId:   drID,
		Height: dr.Height,
	}, nil
}
