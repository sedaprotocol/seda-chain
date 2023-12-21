package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/randomness/types"
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

func (k msgServer) NewSeed(goCtx context.Context, msg *types.MsgNewSeed) (*types.MsgNewSeedResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.SetSeed(sdkCtx, msg.Beta)

	// TO-DO event?

	return &types.MsgNewSeedResponse{}, nil
}
