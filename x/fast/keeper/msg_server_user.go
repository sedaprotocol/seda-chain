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

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

func (m msgServer) AddUser(goCtx context.Context, msg *types.MsgAddUser) (*types.MsgAddUserResponse, error) {
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

	adminAddr, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", msg.AdminAddress)
	}
	if fastClient.AdminAddress != msg.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized admin; expected %s, got %s", fastClient.AdminAddress, msg.AdminAddress)
	}

	hasFastUser, err := m.HasFastUser(ctx, fastClient.Id, msg.UserId)
	if err != nil {
		return nil, err
	}
	if hasFastUser {
		return nil, types.ErrUserAlreadyExists.Wrapf("user already exists for %s", msg.UserId)
	}

	fastUser := types.FastUser{
		UserId:  msg.UserId,
		Credits: msg.InitialCredits,
	}

	events := sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddUser,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeFastClientPubKey, msg.FastClientPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
			sdk.NewAttribute(types.AttributeInitialCredits, msg.InitialCredits.String()),
		),
		formatFastUserEvent(msg.FastClientPublicKey, fastClient.Id, fastUser),
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

		fastClient.Balance = fastClient.Balance.Add(msg.InitialCredits)

		err = m.SetFastClient(ctx, pubKeyBytes, fastClient)
		if err != nil {
			return nil, err
		}

		events = append(events, formatFastClientEvent(fastClient))
	}

	err = m.SetFastUser(ctx, fastClient.Id, msg.UserId, fastUser)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(events)

	return &types.MsgAddUserResponse{}, nil
}

func (m msgServer) RemoveUser(goCtx context.Context, msg *types.MsgRemoveUser) (*types.MsgRemoveUserResponse, error) {
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

	if fastClient.AdminAddress != msg.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized admin; expected %s, got %s", fastClient.AdminAddress, msg.AdminAddress)
	}

	fastUser, err := m.GetFastUser(ctx, fastClient.Id, msg.UserId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("user does not exist for %s", msg.UserId)
		}
		return nil, err
	}

	events := sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveUser,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeFastClientPubKey, msg.FastClientPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
		),
	}

	if fastUser.Credits.IsPositive() {
		fastClient.UsedCredits = fastClient.UsedCredits.Add(fastUser.Credits)

		err = m.SetFastClient(ctx, pubKeyBytes, fastClient)
		if err != nil {
			return nil, err
		}

		events = append(events, formatFastClientEvent(fastClient))
	}

	err = m.DeleteFastUser(ctx, fastClient.Id, msg.UserId)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(events)

	return &types.MsgRemoveUserResponse{}, nil
}

func (m msgServer) TopUpUser(goCtx context.Context, msg *types.MsgTopUpUser) (*types.MsgTopUpUserResponse, error) {
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

	fastUser, err := m.GetFastUser(ctx, fastClient.Id, msg.UserId)
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

	fastClient.Balance = fastClient.Balance.Add(msg.Amount)

	err = m.SetFastClient(ctx, pubKeyBytes, fastClient)
	if err != nil {
		return nil, err
	}

	fastUser.Credits = fastUser.Credits.Add(msg.Amount)

	err = m.SetFastUser(ctx, fastClient.Id, msg.UserId, fastUser)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTopUpUser,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeFastClientPubKey, msg.FastClientPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
			sdk.NewAttribute(types.AttributeCredits, msg.Amount.String()),
		),
		formatFastUserEvent(msg.FastClientPublicKey, fastClient.Id, fastUser),
		formatFastClientEvent(fastClient),
	})

	return &types.MsgTopUpUserResponse{}, nil
}

func (m msgServer) ExpireUserCredits(goCtx context.Context, msg *types.MsgExpireUserCredits) (*types.MsgExpireUserCreditsResponse, error) {
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

	if fastClient.AdminAddress != msg.AdminAddress {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("unauthorized admin; expected %s, got %s", fastClient.AdminAddress, msg.AdminAddress)
	}

	fastUser, err := m.GetFastUser(ctx, fastClient.Id, msg.UserId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("user does not exist for %s", msg.UserId)
		}
		return nil, err
	}

	if fastUser.Credits.LT(msg.Amount) {
		return nil, types.ErrInsufficientCredits.Wrapf("user does not have enough credits; requested %s, available %s", msg.Amount.String(), fastUser.Credits.String())
	}

	fastClient.UsedCredits = fastClient.UsedCredits.Add(msg.Amount)

	err = m.SetFastClient(ctx, pubKeyBytes, fastClient)
	if err != nil {
		return nil, err
	}

	fastUser.Credits = fastUser.Credits.Sub(msg.Amount)

	err = m.SetFastUser(ctx, fastClient.Id, msg.UserId, fastUser)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeExpireUserCredits,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeFastClientPubKey, msg.FastClientPublicKey),
			sdk.NewAttribute(types.AttributeUserID, msg.UserId),
		),
		formatFastClientEvent(fastClient),
		formatFastUserEvent(msg.FastClientPublicKey, fastClient.Id, fastUser),
	})

	return &types.MsgExpireUserCreditsResponse{}, nil
}

func formatFastUserEvent(fastClientPubKey string, fastClientID uint64, user types.FastUser) sdk.Event {
	return sdk.NewEvent(
		types.EventTypeUser,
		sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClientID, 10)),
		sdk.NewAttribute(types.AttributeFastClientPubKey, fastClientPubKey),
		sdk.NewAttribute(types.AttributeUserID, user.UserId),
		sdk.NewAttribute(types.AttributeCredits, user.Credits.String()),
	)
}
