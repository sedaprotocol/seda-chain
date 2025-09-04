package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) RegisterFastClient(goCtx context.Context, msg *types.MsgRegisterFastClient) (*types.MsgRegisterFastClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	pubKeyBytes, err := hex.DecodeString(msg.PublicKey)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid hex in pubkey: %s", msg.PublicKey)
	}

	hasFastClient, err := m.HasFastClient(ctx, pubKeyBytes)
	if err != nil {
		return nil, err
	}
	if hasFastClient {
		return nil, types.ErrFastClientAlreadyExists
	}

	fastClientInput := types.FastClientInput{
		OwnerAddress: msg.OwnerAddress,
		AdminAddress: msg.AdminAddress,
		Address:      msg.Address,
		PublicKey:    pubKeyBytes,
		Memo:         msg.Memo,
		Balance:      math.NewInt(0),
		UsedCredits:  math.NewInt(0),
	}

	fastClient, err := m.CreateFastClient(ctx, pubKeyBytes, fastClientInput)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRegisterFastClient,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
		),
		formatFastClientEvent(fastClient),
	})

	return &types.MsgRegisterFastClientResponse{}, nil
}

func (m msgServer) EditFastClient(goCtx context.Context, msg *types.MsgEditFastClient) (*types.MsgEditFastClientResponse, error) {
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

	// Update fields if provided
	if msg.NewAdminAddress != types.DoNotModifyField {
		fastClient.AdminAddress = msg.NewAdminAddress
	}

	if msg.NewAddress != types.DoNotModifyField {
		fastClient.Address = msg.NewAddress
	}

	if msg.NewMemo != types.DoNotModifyField {
		fastClient.Memo = msg.NewMemo
	}

	if msg.NewPublicKey != types.DoNotModifyField {
		pubKeyBytes, err := hex.DecodeString(msg.NewPublicKey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "invalid hex in new public key: %s", msg.NewPublicKey)
		}

		// Delete old storage
		err = m.DeleteFastClient(ctx, fastClient.PublicKey)
		if err != nil {
			return nil, err
		}

		// Set new public key
		fastClient.PublicKey = pubKeyBytes
	}

	err = m.SetFastClient(ctx, fastClient.PublicKey, fastClient)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateFastClient,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
		),
		formatFastClientEvent(fastClient),
	})

	return &types.MsgEditFastClientResponse{}, nil
}

func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", msg.Authority)
	}
	if m.GetAuthority() != msg.Authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized authority; expected %s, got %s", m.GetAuthority(), msg.Authority)
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func formatFastClientEvent(fastClient types.FastClient) sdk.Event {
	return sdk.NewEvent(
		types.EventTypeFastClient,
		sdk.NewAttribute(types.AttributeFastClientPubKey, hex.EncodeToString(fastClient.PublicKey)),
		sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
		sdk.NewAttribute(types.AttributeOwnerAddress, fastClient.OwnerAddress),
		sdk.NewAttribute(types.AttributeAdminAddress, fastClient.AdminAddress),
		sdk.NewAttribute(types.AttributeAddress, fastClient.Address),
		sdk.NewAttribute(types.AttributeMemo, fastClient.Memo),
		sdk.NewAttribute(types.AttributeBalance, fastClient.Balance.String()),
		sdk.NewAttribute(types.AttributeUsedCredits, fastClient.UsedCredits.String()),
	)
}
