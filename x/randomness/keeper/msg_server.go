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
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TO-DO spam prevention?

	k.Keeper.SetSeed(ctx, msg.Beta)

	// TO-DO event?
	// err = ctx.EventManager().EmitTypedEvent(
	// 	&types.EventNewSeed{
	// 		Hash:     hashString,
	// 		WasmType: msg.WasmType,
	// 		Bytecode: msg.Wasm,
	// 	})
	// if err != nil {
	// 	return nil, err
	// }

	return &types.MsgNewSeedResponse{}, nil
}
