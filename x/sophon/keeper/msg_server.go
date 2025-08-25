package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) RegisterSophon(_ context.Context, _ *types.MsgRegisterSophon) (*types.MsgRegisterSophonResponse, error) {
	panic("not implemented")
}

func (m msgServer) EditSophon(_ context.Context, _ *types.MsgEditSophon) (*types.MsgEditSophonResponse, error) {
	panic("not implemented")
}

func (m msgServer) TransferOwnership(_ context.Context, _ *types.MsgTransferOwnership) (*types.MsgTransferOwnershipResponse, error) {
	panic("not implemented")
}

func (m msgServer) AcceptOwnership(_ context.Context, _ *types.MsgAcceptOwnership) (*types.MsgAcceptOwnershipResponse, error) {
	panic("not implemented")
}

func (m msgServer) CancelOwnershipTransfer(_ context.Context, _ *types.MsgCancelOwnershipTransfer) (*types.MsgCancelOwnershipTransferResponse, error) {
	panic("not implemented")
}

func (m msgServer) AddUser(_ context.Context, _ *types.MsgAddUser) (*types.MsgAddUserResponse, error) {
	panic("not implemented")
}

func (m msgServer) TopUpUser(_ context.Context, _ *types.MsgTopUpUser) (*types.MsgTopUpUserResponse, error) {
	panic("not implemented")
}

func (m msgServer) ExpireCredits(_ context.Context, _ *types.MsgExpireCredits) (*types.MsgExpireCreditsResponse, error) {
	panic("not implemented")
}

func (m msgServer) SettleCredits(_ context.Context, _ *types.MsgSettleCredits) (*types.MsgSettleCreditsResponse, error) {
	panic("not implemented")
}

func (m msgServer) SubmitReports(_ context.Context, _ *types.MsgSubmitReports) (*types.MsgSubmitReportsResponse, error) {
	panic("not implemented")
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
