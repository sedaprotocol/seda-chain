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

func (m msgServer) RemoveUser(_ context.Context, _ *types.MsgRemoveUser) (*types.MsgRemoveUserResponse, error) {
	panic("not implemented")
}

func (m msgServer) TopUpUser(_ context.Context, _ *types.MsgTopUpUser) (*types.MsgTopUpUserResponse, error) {
	panic("not implemented")
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
