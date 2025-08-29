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

func (m msgServer) SettleCredits(goCtx context.Context, msg *types.MsgSettleCredits) (*types.MsgSettleCreditsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	adminAddr, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return nil, err
	}

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
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("invalid admin address; requested %s, available %s", msg.AdminAddress, sophonInfo.AdminAddress)
	}

	if sophonInfo.UsedCredits.LT(msg.Amount) {
		return nil, types.ErrInsufficientCredits.Wrapf("unable to settle credits; requested %s, available %s", msg.Amount.String(), sophonInfo.UsedCredits.String())
	}

	if sophonInfo.Balance.LT(msg.Amount) {
		return nil, types.ErrInsufficientBalance.Wrapf("unable to settle credits; requested %s, available %s", msg.Amount.String(), sophonInfo.Balance.String())
	}

	sophonInfo.UsedCredits = sophonInfo.UsedCredits.Sub(msg.Amount)
	sophonInfo.Balance = sophonInfo.Balance.Sub(msg.Amount)

	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	switch msg.SettleType {
	case types.SETTLE_TYPE_BURN:
		err = m.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, msg.Amount)))
		if err != nil {
			return nil, err
		}
	case types.SETTLE_TYPE_WITHDRAW:
		err = m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, adminAddr, sdk.NewCoins(sdk.NewCoin(denom, msg.Amount)))
		if err != nil {
			return nil, err
		}
	}

	err = m.SetSophonInfo(ctx, pubKeyBytes, sophonInfo)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSettleCredits,
			sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
			sdk.NewAttribute(types.AttributeSophonPubKey, msg.SophonPublicKey),
			sdk.NewAttribute(types.AttributeSettleType, msg.SettleType.String()),
			sdk.NewAttribute(types.AttributeCredits, msg.Amount.String()),
		),
		createSophonEvent(sophonInfo),
	})

	return &types.MsgSettleCreditsResponse{}, nil
}

func (m msgServer) SubmitReports(_ context.Context, _ *types.MsgSubmitReports) (*types.MsgSubmitReportsResponse, error) {
	panic("not implemented")
}
