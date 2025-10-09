package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (m msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.Stake(ctx, msg, false)
	if err != nil {
		return nil, err
	}
	return &types.MsgStakeResponse{}, nil
}

func (m msgServer) LegacyStake(goCtx context.Context, msg *types.MsgLegacyStake) (*types.MsgLegacyStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.Stake(ctx, msg, true)
	if err != nil {
		return nil, err
	}
	return &types.MsgLegacyStakeResponse{}, nil
}

func (m msgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.Unstake(ctx, msg, false)
	if err != nil {
		return nil, err
	}
	return &types.MsgUnstakeResponse{}, nil
}

func (m msgServer) LegacyUnstake(goCtx context.Context, msg *types.MsgLegacyUnstake) (*types.MsgLegacyUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.Unstake(ctx, msg, true)
	if err != nil {
		return nil, err
	}
	return &types.MsgLegacyUnstakeResponse{}, nil
}

func (m msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.Withdraw(ctx, msg, false)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawResponse{}, nil
}

func (m msgServer) LegacyWithdraw(goCtx context.Context, msg *types.MsgLegacyWithdraw) (*types.MsgLegacyWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := m.Keeper.Withdraw(ctx, msg, true)
	if err != nil {
		return nil, err
	}
	return &types.MsgLegacyWithdrawResponse{}, nil
}
