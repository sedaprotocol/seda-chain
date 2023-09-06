package keeper

import (
	"context"
	"encoding/hex"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

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

	wasm := types.NewWasm(unzipped, msg.WasmType)
	k.Keeper.SetDataRequestWasm(ctx, wasm)

	hashString := hex.EncodeToString(wasm.Hash)
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

	wasm := types.NewWasm(unzipped, msg.WasmType)
	k.Keeper.SetOverlayWasm(ctx, wasm)

	hashString := hex.EncodeToString(wasm.Hash)
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
