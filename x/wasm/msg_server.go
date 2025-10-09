package wasm

import (
	"context"
	"encoding/json"

	"github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
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

// ExecuteContract overrides the default ExecuteContract method.
func (m msgServer) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid contract address: %s", msg.Contract)
	}
	coreContractAddr, err := m.WasmStorageKeeper.GetCoreContractAddr(ctx)
	if err != nil {
		return nil, err
	}
	if contractAddr.String() != coreContractAddr.String() {
		return m.MsgServer.ExecuteContract(goCtx, msg)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	bondDenom, err := m.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	// Encode and dispatch.
	var coreContractMsg *CoreContractMsg
	err = json.Unmarshal(msg.Msg, &coreContractMsg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("failed to unmarshal core contract message: %v", err)
	}

	var sdkMsg sdk.Msg
	switch {
	case coreContractMsg.AddToAllowList != nil:
		sdkMsg, err = coreContractMsg.AddToAllowList.EncodeToSdkMsg(msg.Sender)

	case coreContractMsg.Stake != nil:
		_, stake := msg.Funds.Find(bondDenom)
		sdkMsg, err = coreContractMsg.Stake.EncodeToSdkMsg(msg.Sender, stake)

	case coreContractMsg.Unstake != nil:
		sdkMsg, err = coreContractMsg.Unstake.EncodeToSdkMsg(msg.Sender)

	case coreContractMsg.Withdraw != nil:
		sdkMsg, err = coreContractMsg.Withdraw.EncodeToSdkMsg(msg.Sender)

	case coreContractMsg.PostDataRequest != nil:
		_, funds := msg.Funds.Find(bondDenom)
		sdkMsg, err = coreContractMsg.PostDataRequest.EncodeToSdkMsg(msg.Sender, funds)

	case coreContractMsg.CommitDataResult != nil:
		sdkMsg, err = coreContractMsg.CommitDataResult.EncodeToSdkMsg(msg.Sender)

	case coreContractMsg.RevealDataResult != nil:
		sdkMsg, err = coreContractMsg.RevealDataResult.EncodeToSdkMsg(msg.Sender)

	default:
		// TODO Subject to change
		return m.MsgServer.ExecuteContract(goCtx, msg)
	}
	if err != nil {
		return nil, err
	}

	handler := m.router.Handler(sdkMsg)
	if handler == nil {
		return nil, sdkerrors.ErrUnknownRequest.Wrapf("failed to find handler for message type %T", sdkMsg)
	}

	result, err := handler(ctx, sdkMsg)
	if err != nil {
		return nil, err
	}

	returnData := result.Data
	if coreContractMsg.PostDataRequest != nil {
		// Convert proto-encoded response to JSON.
		var res coretypes.MsgPostDataRequestResponse
		if err := res.Unmarshal(returnData); err != nil {
			return nil, err
		}
		contractRes := &PostRequestResponsePayload{
			DrID: res.DrID,
			//nolint:gosec // G115: Block height is never negative.
			Height: uint64(res.Height),
		}
		returnData, err = json.Marshal(contractRes)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgExecuteContractResponse{
		Data: returnData,
	}, nil
}
