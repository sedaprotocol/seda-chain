package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"slices"
	"strconv"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

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

func (m msgServer) SubmitReports(goCtx context.Context, msg *types.MsgSubmitReports) (*types.MsgSubmitReportsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sophonInfo, err := m.GetSophonInfo(ctx, msg.SophonPublicKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("sophon not found for %s", msg.SophonPublicKey)
		}
		return nil, err
	}

	sophonPubKeyHex := hex.EncodeToString(msg.SophonPublicKey)

	if sophonInfo.Address != msg.Address {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("invalid address; received %s, required %s", msg.Address, sophonInfo.Address)
	}

	events := sdk.Events{}

	dataProxyCredits := make(map[string]math.Int)
	totalUsedCredits := math.NewInt(0)

	// There can be at most one report per user.
	for _, report := range msg.Reports {
		sophonUser, err := m.GetSophonUser(ctx, sophonInfo.Id, report.UserId)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return nil, sdkerrors.ErrNotFound.Wrapf("user not found for %s", report.UserId)
			}
			return nil, err
		}

		if sophonUser.Credits.LT(report.UsedCredits) {
			return nil, types.ErrInsufficientCredits.Wrapf("unable to submit report for user %s; requested %s, available %s", report.UserId, report.UsedCredits.String(), sophonUser.Credits.String())
		}

		// UsedCredits includes the data proxy credits so we don't need to check those separately.
		sophonUser.Credits = sophonUser.Credits.Sub(report.UsedCredits)
		totalUsedCredits = totalUsedCredits.Add(report.UsedCredits)

		err = m.SetSophonUser(ctx, sophonInfo.Id, report.UserId, sophonUser)
		if err != nil {
			return nil, err
		}

		events = append(events,
			sdk.NewEvent(
				types.EventTypeUseUserCredits,
				sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
				sdk.NewAttribute(types.AttributeSophonPubKey, sophonPubKeyHex),
				sdk.NewAttribute(types.AttributeUserID, report.UserId),
				sdk.NewAttribute(types.AttributeCredits, report.UsedCredits.String()),
			),
			createUserEvent(sophonPubKeyHex, sophonInfo.Id, sophonUser),
		)

		for _, dataProxyReport := range report.DataProxyReports {
			if dataProxyCredits[dataProxyReport.DataProxyPubKey].IsNil() {
				dataProxyCredits[dataProxyReport.DataProxyPubKey] = math.NewInt(0)
			}

			dataProxyCredits[dataProxyReport.DataProxyPubKey] = dataProxyCredits[dataProxyReport.DataProxyPubKey].Add(dataProxyReport.Price.Mul(math.NewIntFromUint64(dataProxyReport.Amount)))
		}

	}

	denom, err := m.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, err
	}

	sortedDataProxies, err := sortDataProxies(dataProxyCredits)
	if err != nil {
		return nil, err
	}

	totalDataProxyCredits := math.NewInt(0)
	// Transfer the data proxy credits to the data proxy.
	for _, dataProxy := range sortedDataProxies {
		dataProxyConfig, err := m.dataProxyKeeper.GetDataProxyConfig(ctx, dataProxy.PubKey)
		if err != nil {
			// We skip the data proxy if it's not registered
			if errors.Is(err, collections.ErrNotFound) {
				events = append(events, sdk.NewEvent(
					types.EventTypeUnregisteredDataProxy,
					sdk.NewAttribute(types.AttributeDataProxyPubKey, dataProxy.PubKeyHex),
				))
				continue
			}
			return nil, err
		}

		payoutAddress, err := sdk.AccAddressFromBech32(dataProxyConfig.PayoutAddress)
		if err != nil {
			// We skip the data proxy if the payout address is invalid, which shouldn't happen as it's validated when
			// registering or updating the data proxy.
			events = append(events, sdk.NewEvent(
				types.EventTypeInvalidDataProxyPayoutAddress,
				sdk.NewAttribute(types.AttributeDataProxyPubKey, dataProxy.PubKeyHex),
				sdk.NewAttribute(types.AttributeDataProxyPayoutAddress, dataProxyConfig.PayoutAddress),
				sdk.NewAttribute(types.AttributeAmount, dataProxy.Amount.String()),
			))
			continue
		}

		err = m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, payoutAddress, sdk.NewCoins(sdk.NewCoin(denom, dataProxy.Amount)))
		if err != nil {
			return nil, err
		}

		events = append(events, sdk.NewEvent(
			types.EventTypeDataProxyPayout,
			sdk.NewAttribute(types.AttributeDataProxyPubKey, dataProxy.PubKeyHex),
			sdk.NewAttribute(types.AttributeDataProxyPayoutAddress, dataProxyConfig.PayoutAddress),
			sdk.NewAttribute(types.AttributeAmount, dataProxy.Amount.String()),
		))

		totalDataProxyCredits = totalDataProxyCredits.Add(dataProxy.Amount)
	}

	// The Sophon should have paid out the data proxy credits, so we subtract them from the balance.
	sophonInfo.Balance = sophonInfo.Balance.Sub(totalDataProxyCredits)

	// The total credits includes the data proxy credits, but as those get resolved immediately we should
	// not add them to the used credits of the Sophon.
	totalSophonCreditsUsed := totalUsedCredits.Sub(totalDataProxyCredits)
	sophonInfo.UsedCredits = sophonInfo.UsedCredits.Add(totalSophonCreditsUsed)

	err = m.SetSophonInfo(ctx, msg.SophonPublicKey, sophonInfo)
	if err != nil {
		return nil, err
	}

	events = append(events, sdk.NewEvent(
		types.EventTypeSubmitReports,
		sdk.NewAttribute(types.AttributeSophonID, strconv.FormatUint(sophonInfo.Id, 10)),
		sdk.NewAttribute(types.AttributeSophonPubKey, sophonPubKeyHex),
	), createSophonEvent(sophonInfo))

	ctx.EventManager().EmitEvents(events)

	return &types.MsgSubmitReportsResponse{}, nil
}

type sortableDataProxy struct {
	PubKey    []byte
	PubKeyHex string
	Amount    math.Int
}

// Sort the data proxies by pubkey
func sortDataProxies(dataProxyCredits map[string]math.Int) ([]sortableDataProxy, error) {
	sortedDataProxies := make([]sortableDataProxy, 0, len(dataProxyCredits))
	for dataProxyPubKey, amount := range dataProxyCredits {
		dataProxyPubKeyBytes, err := hex.DecodeString(dataProxyPubKey)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "invalid hex in pubkey: %s", dataProxyPubKey)
		}
		sortedDataProxies = append(sortedDataProxies, sortableDataProxy{
			PubKey:    dataProxyPubKeyBytes,
			PubKeyHex: dataProxyPubKey,
			Amount:    amount,
		})
	}
	slices.SortFunc(sortedDataProxies, func(a, b sortableDataProxy) int {
		return bytes.Compare(a.PubKey, b.PubKey)
	})
	return sortedDataProxies, nil
}
