package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
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

	data := msg.Data

	var err error
	if ioutils.IsGzip(data) {
		data, err = ioutils.Uncompress(data, int64(800*1024))
		if err != nil {
			return nil, err
		}
	}

	hash := crypto.Keccak256(data)
	if hash == nil {
		return nil, fmt.Errorf("hash of the data is nil")
	}
	k.Keeper.SetData(ctx, data, hash)

	hashString := hex.EncodeToString(hash)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStore,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeHash, hashString),
		),
	)
	return &types.MsgStoreResponse{
		Hash: hashString,
	}, nil
}
