package keeper

import (
	"context"
	"encoding/hex"

	"cosmossdk.io/math"

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

func (m msgServer) RegisterSophon(goCtx context.Context, msg *types.MsgRegisterSophon) (*types.MsgRegisterSophonResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	pubKeyBytes, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", msg.PublicKey)
	}

	hasSophon, err := m.HasSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		return nil, err
	}
	if hasSophon {
		return nil, types.ErrAlreadyExists
	}

	sophonInfo := types.SophonInputs{
		OwnerAddress: msg.OwnerAddress,
		AdminAddress: msg.AdminAddress,
		Address:      msg.Address,
		PublicKey:    pubKeyBytes,
		Memo:         msg.Memo,
		Balance:      math.NewInt(0),
		UsedCredits:  math.NewInt(0),
	}

	_, err = m.Keeper.CreateSophonInfo(ctx, pubKeyBytes, sophonInfo)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRegisterSophon,
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.PublicKey),
			sdk.NewAttribute(types.AttributeOwnerAddress, msg.OwnerAddress),
			sdk.NewAttribute(types.AttributeAdminAddress, msg.AdminAddress),
			sdk.NewAttribute(types.AttributeAddress, msg.Address),
			sdk.NewAttribute(types.AttributeMemo, msg.Memo),
			sdk.NewAttribute(types.AttributeBalance, sophonInfo.Balance.String()),
			sdk.NewAttribute(types.AttributeUsedCredits, sophonInfo.UsedCredits.String()),
		),
	})

	return &types.MsgRegisterSophonResponse{}, nil
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
