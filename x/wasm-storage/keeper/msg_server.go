package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type msgServer struct {
	Keeper
}
type EventStoreDataRequestWasmWrapper struct {
	*types.EventStoreDataRequestWasm
}

// MarshalJSON customizes the JSON encoding of the type that implements it
func (e EventStoreDataRequestWasmWrapper) MarshalJSON() ([]byte, error) {
	// avoid infinite recursion when calling json.Marshal
	type Alias types.EventStoreDataRequestWasm

	return json.Marshal(&struct {
		Hash json.RawMessage `json:"hash"`
		*Alias
	}{
		Hash:  json.RawMessage(`"` + e.Hash + `"`),   // wrap the raw json value in double quotes
		Alias: (*Alias)(e.EventStoreDataRequestWasm), // cast to embedded type
	})
}

type EventStoreOverlayWasmWrapper struct {
	*types.EventStoreOverlayWasm
}

func (e EventStoreOverlayWasmWrapper) MarshalJSON() ([]byte, error) {
	type Alias types.EventStoreOverlayWasm
	return json.Marshal(&struct {
		Hash json.RawMessage `json:"hash"`
		*Alias
	}{
		Hash:  json.RawMessage(`"` + e.Hash + `"`),
		Alias: (*Alias)(e.EventStoreOverlayWasm),
	})
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) StoreDataRequestWasm(goCtx context.Context, msg *types.MsgStoreDataRequestWasm) (*types.MsgStoreDataRequestWasmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	unzipped, err := unzipWasm(msg.Wasm)
	if err != nil {
		return nil, err
	}
	wasm := types.NewWasm(unzipped, msg.WasmType, ctx.BlockTime())
	exists, _ := m.Keeper.HasDataRequestWasm(ctx, wasm)
	if exists {
		return nil, fmt.Errorf("data Request Wasm with given hash already exists")
	}
	err = m.Keeper.SetDataRequestWasm(ctx, wasm)
	if err != nil {
		return nil, err
	}

	hashString := hex.EncodeToString(wasm.Hash)

	err = ctx.EventManager().EmitTypedEvent(
		&EventStoreDataRequestWasmWrapper{
			EventStoreDataRequestWasm: &types.EventStoreDataRequestWasm{
				Hash:     hashString,
				WasmType: msg.WasmType,
				Bytecode: msg.Wasm,
			},
		})
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreDataRequestWasmResponse{
		Hash: hashString,
	}, nil
}

func (m msgServer) StoreOverlayWasm(goCtx context.Context, msg *types.MsgStoreOverlayWasm) (*types.MsgStoreOverlayWasmResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Sender != m.authority {
		return nil, fmt.Errorf("invalid authority %s", msg.Sender)
	}

	unzipped, err := unzipWasm(msg.Wasm)
	if err != nil {
		return nil, err
	}
	wasm := types.NewWasm(unzipped, msg.WasmType, ctx.BlockTime())
	exists, _ := m.Keeper.HasOverlayWasm(ctx, wasm)
	if exists {
		return nil, fmt.Errorf("overlay Wasm with given hash already exists")
	}
	err = m.Keeper.SetOverlayWasm(ctx, wasm)
	if err != nil {
		return nil, err
	}

	hashString := hex.EncodeToString(wasm.Hash)
	err = ctx.EventManager().EmitTypedEvent(
		&EventStoreOverlayWasmWrapper{
			EventStoreOverlayWasm: &types.EventStoreOverlayWasm{
				Hash:     hashString,
				WasmType: msg.WasmType,
				Bytecode: msg.Wasm,
			},
		})
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreOverlayWasmResponse{
		Hash: hashString,
	}, nil
}

// InstantiateAndRegisterProxyContract instantiate a new contract with
// a predictable address and updates the Proxy Contract registry.
func (m msgServer) InstantiateAndRegisterProxyContract(goCtx context.Context, msg *types.MsgInstantiateAndRegisterProxyContract) (*types.MsgInstantiateAndRegisterProxyContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var adminAddr sdk.AccAddress
	var err error
	if msg.Admin != "" {
		if adminAddr, err = sdk.AccAddressFromBech32(msg.Admin); err != nil {
			return nil, fmt.Errorf("invalid admin address: %s", err)
		}
	}

	contractAddr, _, err := m.wasmKeeper.Instantiate2(ctx, msg.CodeID, adminAddr, adminAddr, msg.Msg, msg.Label, msg.Funds, msg.Salt, msg.FixMsg)
	if err != nil {
		return nil, err
	}

	// update Proxy Contract registry
	err = m.ProxyContractRegistry.Set(ctx, contractAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgInstantiateAndRegisterProxyContractResponse{
		ContractAddress: contractAddr.String(),
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
		return nil, err
	}
	return unzipped, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// validate authority
	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, fmt.Errorf("invalid authority address: %s", err)
	}

	if m.GetAuthority() != req.Authority {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.GetAuthority(), req.Authority)
	}

	// validate params
	if err := req.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Params.Set(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
