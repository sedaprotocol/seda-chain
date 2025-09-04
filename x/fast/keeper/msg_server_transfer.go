package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

func (m msgServer) TransferOwnership(goCtx context.Context, msg *types.MsgTransferOwnership) (*types.MsgTransferOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.FastClientPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.FastClientPublicKey)
	}

	fastClient, err := m.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("fast client not found for %s", msg.FastClientPublicKey)
		}
		return nil, err
	}

	if fastClient.OwnerAddress != msg.OwnerAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized owner; expected %s, got %s", fastClient.OwnerAddress, msg.OwnerAddress)
	}

	newOwnerAddress, err := sdk.AccAddressFromBech32(msg.NewOwnerAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", msg.NewOwnerAddress)
	}

	err = m.DeleteFastTransfer(ctx, fastClient.Id)
	if err != nil {
		return nil, err
	}

	err = m.SetFastTransfer(ctx, fastClient.Id, newOwnerAddress)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransferOwnership,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeOwnerAddress, msg.OwnerAddress),
			sdk.NewAttribute(types.AttributeNewOwnerAddress, msg.NewOwnerAddress),
		),
	})

	return &types.MsgTransferOwnershipResponse{}, nil
}

func (m msgServer) AcceptOwnership(goCtx context.Context, msg *types.MsgAcceptOwnership) (*types.MsgAcceptOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.FastClientPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.FastClientPublicKey)
	}

	fastClient, err := m.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("fast client not found for %s", msg.FastClientPublicKey)
		}
		return nil, err
	}

	newOwnerAddress, err := sdk.AccAddressFromBech32(msg.NewOwnerAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid new owner address: %s", msg.NewOwnerAddress)
	}

	transferAddress, err := m.GetFastTransfer(ctx, fastClient.Id)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	if transferAddress == nil || !bytes.Equal(transferAddress, newOwnerAddress) {
		return nil, sdkerrors.ErrNotFound.Wrapf("there is no pending ownership transfer for %s", msg.NewOwnerAddress)
	}

	err = m.DeleteFastTransfer(ctx, fastClient.Id)
	if err != nil {
		return nil, err
	}

	fastClient.OwnerAddress = msg.NewOwnerAddress
	err = m.SetFastClient(ctx, fastClient.PublicKey, fastClient)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAcceptOwnership,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeOwnerAddress, msg.NewOwnerAddress),
		),
		formatFastClientEvent(fastClient),
	})

	return &types.MsgAcceptOwnershipResponse{}, nil
}

func (m msgServer) CancelOwnershipTransfer(goCtx context.Context, msg *types.MsgCancelOwnershipTransfer) (*types.MsgCancelOwnershipTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.FastClientPublicKey)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", msg.FastClientPublicKey)
	}

	fastClient, err := m.GetFastClient(ctx, pubKeyBytes)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("fast client not found for %s", msg.FastClientPublicKey)
		}
		return nil, err
	}

	if fastClient.OwnerAddress != msg.OwnerAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized owner; expected %s, got %s", fastClient.OwnerAddress, msg.OwnerAddress)
	}

	transferAddress, err := m.GetFastTransfer(ctx, fastClient.Id)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	if transferAddress == nil {
		return nil, sdkerrors.ErrNotFound.Wrapf("there is no pending ownership transfer for %s", msg.FastClientPublicKey)
	}

	err = m.DeleteFastTransfer(ctx, fastClient.Id)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCancelOwnershipTransfer,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
		),
	})

	return &types.MsgCancelOwnershipTransferResponse{}, nil
}
