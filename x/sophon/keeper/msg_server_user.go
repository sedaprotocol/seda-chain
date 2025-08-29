package keeper

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (m msgServer) AddUser(goCtx context.Context, msg *types.MsgAddUser) (*types.MsgAddUserResponse, error) {
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

	adminAddr, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", msg.AdminAddress)
	}
	if sophonInfo.AdminAddress != msg.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized admin; expected %s, got %s", sophonInfo.AdminAddress, msg.AdminAddress)
	}

	hasSophonUser, err := m.HasSophonUser(ctx, sophonInfo.Id, msg.UserId)
	if err != nil {
		return nil, err
	}
	if hasSophonUser {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("user already exists for %s", msg.UserId)
	}

	sophonUser := types.SophonUser{
		UserId:  msg.UserId,
		Credits: msg.InitialCredits,
	}

	events := sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddUser,
			sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
			sdk.NewAttribute(types.AttributeInitialCredits, msg.InitialCredits.String()),
		),
		createUserEvent(msg.SophonPublicKey, sophonInfo.Id, sophonUser),
	}

	if msg.InitialCredits.IsPositive() {
		denom, err := m.stakingKeeper.BondDenom(ctx)
		if err != nil {
			return nil, err
		}
		err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, adminAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, msg.InitialCredits)))
		if err != nil {
			return nil, err
		}

		sophonInfo.Balance = sophonInfo.Balance.Add(msg.InitialCredits)

		err = m.SetSophonInfo(ctx, pubKeyBytes, sophonInfo)
		if err != nil {
			return nil, err
		}

		events = append(events, createSophonEvent(sophonInfo))
	}

	err = m.SetSophonUser(ctx, sophonInfo.Id, msg.UserId, sophonUser)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(events)

	return &types.MsgAddUserResponse{}, nil
}

func (m msgServer) RemoveUser(goCtx context.Context, msg *types.MsgRemoveUser) (*types.MsgRemoveUserResponse, error) {
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

	if sophonInfo.AdminAddress != msg.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized admin; expected %s, got %s", sophonInfo.AdminAddress, msg.AdminAddress)
	}

	sophonUser, err := m.GetSophonUser(ctx, sophonInfo.Id, msg.UserId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("user does not exist for %s", msg.UserId)
		}
		return nil, err
	}

	events := sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveUser,
			sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
		),
	}

	if sophonUser.Credits.IsPositive() {
		sophonInfo.UsedCredits = sophonInfo.UsedCredits.Add(sophonUser.Credits)

		err = m.SetSophonInfo(ctx, pubKeyBytes, sophonInfo)
		if err != nil {
			return nil, err
		}

		events = append(events, createSophonEvent(sophonInfo))
	}

	err = m.DeleteSophonUser(ctx, sophonInfo.Id, msg.UserId)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(events)

	return &types.MsgRemoveUserResponse{}, nil
}

func (m msgServer) TopUpUser(goCtx context.Context, msg *types.MsgTopUpUser) (*types.MsgTopUpUserResponse, error) {
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

	sophonUser, err := m.GetSophonUser(ctx, sophonInfo.Id, msg.UserId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("user not found for %s", msg.UserId)
		}
		return nil, err
	}

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", msg.Sender)
	}

	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}
	err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, msg.Amount)))
	if err != nil {
		return nil, err
	}

	sophonInfo.Balance = sophonInfo.Balance.Add(msg.Amount)

	err = m.SetSophonInfo(ctx, pubKeyBytes, sophonInfo)
	if err != nil {
		return nil, err
	}

	sophonUser.Credits = sophonUser.Credits.Add(msg.Amount)

	err = m.SetSophonUser(ctx, sophonInfo.Id, msg.UserId, sophonUser)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTopUpUser,
			sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
			sdk.NewAttribute(types.AttributeCredits, msg.Amount.String()),
		),
		createUserEvent(msg.SophonPublicKey, sophonInfo.Id, sophonUser),
		createSophonEvent(sophonInfo),
	})

	return &types.MsgTopUpUserResponse{}, nil
}

func (m msgServer) ExpireCredits(_ context.Context, _ *types.MsgExpireCredits) (*types.MsgExpireCreditsResponse, error) {
	panic("not implemented")
}

func createUserEvent(sophonPubKey string, sophonID uint64, user types.SophonUser) sdk.Event {
	return sdk.NewEvent(
		types.EventTypeUser,
		sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonID, 10)),
		sdk.NewAttribute(types.AttributeSophonPubKey, sophonPubKey),
		sdk.NewAttribute(types.AttributeUserID, user.UserId),
		sdk.NewAttribute(types.AttributeCredits, user.Credits.String()),
	)
}
