package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (m msgServer) TransferOwnership(goCtx context.Context, msg *types.MsgTransferOwnership) (*types.MsgTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.SophonPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.SophonPublicKey)
	}

	sophonInfo, err := m.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("sophon not found for %s", msg.SophonPublicKey)
		}
		return nil, err
	}

	if sophonInfo.OwnerAddress != msg.OwnerAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized owner; expected %s, got %s", sophonInfo.OwnerAddress, msg.OwnerAddress)
	}

	newOwnerAddress, err := sdk.AccAddressFromBech32(msg.NewOwnerAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", msg.NewOwnerAddress)
	}

	err = m.DeleteSophonTransfer(ctx, sophonInfo.Id)
	if err != nil {
		return nil, err
	}

	err = m.SetSophonTransfer(ctx, sophonInfo.Id, newOwnerAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransferOwnership,
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
			sdk.NewAttribute(types.AttributeOwnerAddress, msg.OwnerAddress),
			sdk.NewAttribute(types.AttributeNewOwnerAddress, msg.NewOwnerAddress),
		),
	})

	return &types.MsgTransferOwnershipResponse{}, nil
}

func (m msgServer) AcceptOwnership(goCtx context.Context, msg *types.MsgAcceptOwnership) (*types.MsgAcceptOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.SophonPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.SophonPublicKey)
	}

	sophonInfo, err := m.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("sophon not found for %s", msg.SophonPublicKey)
		}
		return nil, err
	}

	newOwnerAddress, err := sdk.AccAddressFromBech32(msg.NewOwnerAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", msg.NewOwnerAddress)
	}

	transferAddress, err := m.GetSophonTransfer(ctx, sophonInfo.Id)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	if transferAddress == nil || !bytes.Equal(transferAddress, newOwnerAddress) {
		return nil, sdkerrors.ErrNotFound.Wrapf("there is no pending ownership transfer for %s", msg.NewOwnerAddress)
	}

	err = m.DeleteSophonTransfer(ctx, sophonInfo.Id)
	if err != nil {
		return nil, err
	}

	sophonInfo.OwnerAddress = msg.NewOwnerAddress
	err = m.SetSophonInfo(ctx, sophonInfo.PublicKey, sophonInfo)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAcceptOwnership,
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
			sdk.NewAttribute(types.AttributeOwnerAddress, msg.NewOwnerAddress),
		),
		createSophonEvent(types.EventTypeUpdateSophon, sophonInfo),
	})

	return &types.MsgAcceptOwnershipResponse{}, nil
}

func (m msgServer) CancelOwnershipTransfer(goCtx context.Context, msg *types.MsgCancelOwnershipTransfer) (*types.MsgCancelOwnershipTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.SophonPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.SophonPublicKey)
	}

	sophonInfo, err := m.GetSophonInfo(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("sophon not found for %s", msg.SophonPublicKey)
		}
		return nil, err
	}

	if sophonInfo.OwnerAddress != msg.OwnerAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized owner; expected %s, got %s", sophonInfo.OwnerAddress, msg.OwnerAddress)
	}

	transferAddress, err := m.GetSophonTransfer(ctx, sophonInfo.Id)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	if transferAddress == nil {
		return nil, sdkerrors.ErrNotFound.Wrapf("there is no pending ownership transfer for %s", msg.SophonPublicKey)
	}

	err = m.DeleteSophonTransfer(ctx, sophonInfo.Id)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelOwnershipTransfer,
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
		),
	})

	return &types.MsgCancelOwnershipTransferResponse{}, nil
}
