package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// UpdateParams updates the module parameters.
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}
	if m.GetAuthority() != req.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("expected %s, got %s", m.GetAuthority(), req.Authority)
	}

	err = req.Params.Validate()
	if err != nil {
		return nil, err
	}

	err = m.SetParams(ctx, req.Params)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}
