package keeper

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
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

func (m msgServer) AddVrfKey(goCtx context.Context, msg *types.MsgAddVRFKey) (*types.MsgAddVrfKeyResponse, error) {
	if err := msg.ValidateBasic(m.Modules); err != nil {
		return nil, err
	}

	pubKey, err := msg.PublicKey()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.Keeper.PublicKeys.Set(ctx, collections.Join(msg.Application, msg.Name), pubKey); err != nil {
		return nil, err
	}
	return &types.MsgAddVrfKeyResponse{}, nil
}
