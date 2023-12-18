package staking

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

var _ stakingtypes.MsgServer = msgServer{}

type msgServer struct {
	stakingtypes.MsgServer
	accountKeeper types.AccountKeeper
}

func NewMsgServerImpl(keeper *stakingkeeper.Keeper, accKeeper types.AccountKeeper) stakingtypes.MsgServer {
	ms := &msgServer{
		MsgServer:     stakingkeeper.NewMsgServerImpl(keeper),
		accountKeeper: accKeeper,
	}
	return ms
}

func (k msgServer) CreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	// create an account based on consensus key to send NewSeed txs when proposing blocks
	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	addr := sdk.AccAddress(pk.Address().Bytes())
	acc := k.accountKeeper.NewAccountWithAddress(ctx, addr)
	k.accountKeeper.SetAccount(ctx, acc)

	return k.MsgServer.CreateValidator(ctx, msg)
}

func (k msgServer) EditValidator(ctx context.Context, msg *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error) {
	return k.MsgServer.EditValidator(ctx, msg)
}

func (k msgServer) Delegate(ctx context.Context, msg *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	return k.MsgServer.Delegate(ctx, msg)
}

func (k msgServer) BeginRedelegate(ctx context.Context, msg *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error) {
	return k.MsgServer.BeginRedelegate(ctx, msg)
}

func (k msgServer) Undelegate(ctx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	return k.MsgServer.Undelegate(ctx, msg)
}

func (k msgServer) CancelUnbondingDelegation(ctx context.Context, msg *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error) {
	return k.MsgServer.CancelUnbondingDelegation(ctx, msg)
}

func (k msgServer) UpdateParams(ctx context.Context, msg *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error) {
	return k.MsgServer.UpdateParams(ctx, msg)
}
