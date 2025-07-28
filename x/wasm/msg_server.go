package wasm

import (
	"context"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	types.MsgServer
	*Keeper
}

func NewMsgServerImpl(sdkMsgServer types.MsgServer, keeper *Keeper) types.MsgServer {
	ms := &msgServer{
		MsgServer: sdkMsgServer,
		Keeper:    keeper,
	}
	return ms
}

// TODO Override all methods?

// ExecuteContract overrides the default ExecuteContract method.
func (m msgServer) ExecuteContract(ctx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	// if err != nil {
	// 	return nil, errorsmod.Wrap(err, "sender")
	// }
	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid contract address: %s", msg.Contract)
	}

	coreContractAddr, err := m.WasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}
	if contractAddr.String() != coreContractAddr.String() {
		return nil, sdkerrors.ErrUnknownRequest.Wrapf("unsupported contract address")
	}

	// TODO Parse and call the correct x/core message handler.
	var data []byte
	// data, err := m.keeper.execute(ctx, contractAddr, senderAddr, msg.Msg, msg.Funds)
	// if err != nil {
	// 	return nil, err
	// }

	return &types.MsgExecuteContractResponse{
		Data: data,
	}, nil
}
