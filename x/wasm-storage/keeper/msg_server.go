package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/hyperledger/burrow/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
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

func (k msgServer) StoreDataRequestWasm(goCtx context.Context, msg *types.MsgStoreDataRequestWasm) (*types.MsgStoreDataRequestWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// unzip
	var unzipped []byte
	var err error
	if ioutils.IsGzip(msg.Wasm) {
		unzipped, err = ioutils.Uncompress(msg.Wasm, int64(800*1024))
		if err != nil {
			return nil, err
		}
	}

	// // TO-DO: Check if Wasm?
	// if !ioutils.IsWasm(unzipped) {
	// }

	hash := crypto.Keccak256(unzipped)
	if hash == nil {
		return nil, fmt.Errorf("hash of Wasm is nil")
	}
	k.Keeper.SetDataRequestWasm(ctx, unzipped, hash)

	hashString := hex.EncodeToString(hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStore,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeHash, hashString),
		),
	)
	return &types.MsgStoreDataRequestWasmResponse{
		Hash: hashString,
	}, nil
}

func (k msgServer) StoreOverlayWasm(goCtx context.Context, msg *types.MsgStoreOverlayWasm) (*types.MsgStoreOverlayWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// unzip
	var unzipped []byte
	var err error
	if ioutils.IsGzip(msg.Wasm) {
		unzipped, err = ioutils.Uncompress(msg.Wasm, int64(800*1024))
		if err != nil {
			return nil, err
		}
	}

	// // TO-DO: Check if Wasm?
	// if !ioutils.IsWasm(unzipped) {
	// }

	hash := crypto.Keccak256(unzipped)
	if hash == nil {
		return nil, fmt.Errorf("hash of Wasm is nil")
	}
	k.Keeper.SetDataRequestWasm(ctx, unzipped, hash)

	hashString := hex.EncodeToString(hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStore,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeHash, hashString),
		),
	)
	return &types.MsgStoreOverlayWasmResponse{
		Hash: hashString,
	}, nil
}
