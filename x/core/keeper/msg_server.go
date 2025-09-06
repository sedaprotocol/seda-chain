package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/core/types"
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

func (m msgServer) AddToAllowlist(goCtx context.Context, msg *types.MsgAddToAllowlist) (*types.MsgAddToAllowlistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Sender != m.GetAuthority() {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Sender)
	}

	exists, err := m.Allowlist.Has(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, types.ErrAlreadyAllowlisted
	}

	err = m.Allowlist.Set(ctx, msg.PublicKey)
	if err != nil {
		return nil, err
	}

	// TODO: add event
	// Ok(Response::new().add_attribute("action", "add-to-allowlist").add_event(
	// 	Event::new("seda-contract").add_attributes([
	// 		("version", CONTRACT_VERSION.to_string()),
	// 		("identity", self.public_key),
	// 		("action", "allowlist-add".to_string()),
	// 	]),
	// ))
	return &types.MsgAddToAllowlistResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
