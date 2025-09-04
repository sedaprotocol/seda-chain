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

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

func (m msgServer) SettleCredits(goCtx context.Context, msg *types.MsgSettleCredits) (*types.MsgSettleCreditsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	adminAddr, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return nil, err
	}

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
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("invalid admin address; requested %s, available %s", msg.AdminAddress, fastClient.AdminAddress)
	}

	if fastClient.UsedCredits.LT(msg.Amount) {
		return nil, types.ErrInsufficientCredits.Wrapf("unable to settle credits; requested %s, available %s", msg.Amount.String(), fastClient.UsedCredits.String())
	}

	if fastClient.Balance.LT(msg.Amount) {
		return nil, types.ErrInsufficientBalance.Wrapf("unable to settle credits; requested %s, available %s", msg.Amount.String(), fastClient.Balance.String())
	}

	fastClient.UsedCredits = fastClient.UsedCredits.Sub(msg.Amount)
	fastClient.Balance = fastClient.Balance.Sub(msg.Amount)

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

	err = m.SetFastClient(ctx, pubKeyBytes, fastClient)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSettleCredits,
			sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
			sdk.NewAttribute(types.AttributeFastClientPubKey, msg.FastClientPublicKey),
			sdk.NewAttribute(types.AttributeSettleType, msg.SettleType.String()),
			sdk.NewAttribute(types.AttributeCredits, msg.Amount.String()),
		),
		formatFastClientEvent(fastClient),
	})

	return &types.MsgSettleCreditsResponse{}, nil
}

func (m msgServer) SubmitReports(goCtx context.Context, msg *types.MsgSubmitReports) (*types.MsgSubmitReportsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fastClient, err := m.GetFastClient(ctx, msg.FastClientPublicKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrNotFound.Wrapf("fast client not found for %s", msg.FastClientPublicKey)
		}
		return nil, err
	}

	fastClientPubKeyHex := hex.EncodeToString(msg.FastClientPublicKey)

	if fastClient.Address != msg.Address {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("invalid address; received %s, required %s", msg.Address, fastClient.Address)
	}

	events := sdk.Events{}

	dataProxyCredits := make(map[string]math.Int)
	totalUsedCredits := math.NewInt(0)

	// There can be at most one report per user.
	for _, report := range msg.Reports {
		fastUser, err := m.GetFastUser(ctx, fastClient.Id, report.UserId)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return nil, sdkerrors.ErrNotFound.Wrapf("user not found for %s", report.UserId)
			}
			return nil, err
		}

		if fastUser.Credits.LT(report.UsedCredits) {
			return nil, types.ErrInsufficientCredits.Wrapf("unable to submit report for user %s; requested %s, available %s", report.UserId, report.UsedCredits.String(), fastUser.Credits.String())
		}

		// UsedCredits includes the data proxy credits so we don't need to check those separately.
		fastUser.Credits = fastUser.Credits.Sub(report.UsedCredits)
		totalUsedCredits = totalUsedCredits.Add(report.UsedCredits)

		err = m.SetFastUser(ctx, fastClient.Id, report.UserId, fastUser)
		if err != nil {
			return nil, err
		}

		events = append(events,
			sdk.NewEvent(
				types.EventTypeUseUserCredits,
				sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
				sdk.NewAttribute(types.AttributeFastClientPubKey, fastClientPubKeyHex),
				sdk.NewAttribute(types.AttributeUserID, report.UserId),
				sdk.NewAttribute(types.AttributeCredits, report.UsedCredits.String()),
			),
			formatFastUserEvent(fastClientPubKeyHex, fastClient.Id, fastUser),
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

	// The fast client should have paid out the data proxy credits, so we subtract them from the balance.
	fastClient.Balance = fastClient.Balance.Sub(totalDataProxyCredits)

	// The total credits includes the data proxy credits, but as those get resolved immediately we should
	// not add them to the used credits of the fast client.
	totalFastClientCreditsUsed := totalUsedCredits.Sub(totalDataProxyCredits)
	fastClient.UsedCredits = fastClient.UsedCredits.Add(totalFastClientCreditsUsed)

	err = m.SetFastClient(ctx, msg.FastClientPublicKey, fastClient)
	if err != nil {
		return nil, err
	}

	events = append(events, sdk.NewEvent(
		types.EventTypeSubmitReports,
		sdk.NewAttribute(types.AttributeFastClientID, strconv.FormatUint(fastClient.Id, 10)),
		sdk.NewAttribute(types.AttributeFastClientPubKey, fastClientPubKeyHex),
	), formatFastClientEvent(fastClient))

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
