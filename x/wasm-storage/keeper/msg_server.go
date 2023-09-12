package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

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

	unzipped := unzipWasm(msg.Wasm)
	wasm := types.NewWasm(unzipped, msg.WasmType, ctx.BlockTime())
	if k.Keeper.HasDataRequestWasm(ctx, wasm) {
		return nil, fmt.Errorf("Data Request Wasm with given hash already exists")
	}
	k.Keeper.SetDataRequestWasm(ctx, wasm)

	hashString := hex.EncodeToString(wasm.Hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStoreDataRequestWasm,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeHash, hashString),
			sdk.NewAttribute(types.AttributeWasmType, msg.WasmType.String()),
		),
	)
	return &types.MsgStoreDataRequestWasmResponse{
		Hash: hashString,
	}, nil
}

func (k msgServer) StoreOverlayWasm(goCtx context.Context, msg *types.MsgStoreOverlayWasm) (*types.MsgStoreOverlayWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	unzipped := unzipWasm(msg.Wasm)
	wasm := types.NewWasm(unzipped, msg.WasmType, ctx.BlockTime())
	if k.Keeper.HasOverlayWasm(ctx, wasm) {
		return nil, fmt.Errorf("Overlay Wasm with given hash already exists")
	}
	k.Keeper.SetOverlayWasm(ctx, wasm)

	hashString := hex.EncodeToString(wasm.Hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOverlayWasm,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeHash, hashString),
			sdk.NewAttribute(types.AttributeWasmType, msg.WasmType.String()),
		),
	)
	return &types.MsgStoreOverlayWasmResponse{
		Hash: hashString,
	}, nil
}

// unzipWasm unzips a gzipped Wasm into
func unzipWasm(wasm []byte) []byte {
	var unzipped []byte
	var err error
	if ioutils.IsGzip(wasm) {
		unzipped, err = ioutils.Uncompress(wasm, types.MaxWasmSize)
		if err != nil {
			panic(err)
		}
	}
	return unzipped
}
