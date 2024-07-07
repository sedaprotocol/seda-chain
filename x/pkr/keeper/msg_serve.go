package keeper

import (
	"context"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) AddVrfKey(ctx context.Context, key *types.MsgAddVRFKey) (*types.MsgAddVrfKeyResponse, error) {
	//TODO implement me
	panic("implement me")
}
