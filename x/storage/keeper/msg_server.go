package keeper

import (
	"context"
	"fmt"

	"github.com/hyperledger/burrow/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/storage/types"
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

func (k msgServer) Store(goCtx context.Context, msg *types.MsgStore) (*types.MsgStoreResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hash := crypto.Keccak256(msg.Data)
	if hash == nil {
		return nil, fmt.Errorf("hash of the data is nil")
	}
	k.Keeper.SetData(ctx, msg.Data, hash)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStore,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeHash, string(hash)),
		),
	)

	return &types.MsgStoreResponse{
		Hash: string(hash),
	}, nil
}
