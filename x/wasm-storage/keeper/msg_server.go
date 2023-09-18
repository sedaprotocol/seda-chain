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

	unzipped, err := unzipWasm(msg.Wasm)
	if err != nil {
		return nil, err
	}
	wasm := types.NewWasm(unzipped, msg.WasmType, ctx.BlockTime())
	if k.Keeper.HasDataRequestWasm(ctx, wasm) {
		return nil, fmt.Errorf("data Request Wasm with given hash already exists")
	}
	k.Keeper.SetDataRequestWasm(ctx, wasm)

	hashString := hex.EncodeToString(wasm.Hash)

	err = ctx.EventManager().EmitTypedEvent(
		&types.EventStoreDataRequestWasm{
			Hash:     hashString,
			WasmType: msg.WasmType,
		})
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreDataRequestWasmResponse{
		Hash: hashString,
	}, nil
}

func (k msgServer) StoreOverlayWasm(goCtx context.Context, msg *types.MsgStoreOverlayWasm) (*types.MsgStoreOverlayWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Sender != k.authority {
		return nil, fmt.Errorf("invalid authority %s", msg.Sender)
	}

	unzipped, err := unzipWasm(msg.Wasm)
	if err != nil {
		return nil, err
	}
	wasm := types.NewWasm(unzipped, msg.WasmType, ctx.BlockTime())
	if k.Keeper.HasOverlayWasm(ctx, wasm) {
		return nil, fmt.Errorf("overlay Wasm with given hash already exists")
	}
	k.Keeper.SetOverlayWasm(ctx, wasm)

	hashString := hex.EncodeToString(wasm.Hash)
	err = ctx.EventManager().EmitTypedEvent(
		&types.EventStoreOverlayWasm{
			Hash:     hashString,
			WasmType: msg.WasmType,
		})
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreOverlayWasmResponse{
		Hash: hashString,
	}, nil
}

// unzipWasm unzips a gzipped Wasm into
func unzipWasm(wasm []byte) ([]byte, error) {
	var unzipped []byte
	var err error
	if !ioutils.IsGzip(wasm) {
		return nil, fmt.Errorf("wasm is not gzip compressed")
	}
	unzipped, err = ioutils.Uncompress(wasm, types.MaxWasmSize)
	if err != nil {
		return unzipped, nil
	}
	return unzipped, nil
}
