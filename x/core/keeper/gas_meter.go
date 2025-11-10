package keeper

import (
	"encoding/hex"
	"fmt"
	stdmath "math"
	"slices"
	"sort"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// ChargeGasCosts charges gas costs to the escrow funds based on gas meter
// reading. Remaining escrow funds are refunded to the data request poster.
func (k Keeper) ChargeGasCosts(ctx sdk.Context, denom string, tr *TallyResult, minimumStake math.Int, burnRatio math.LegacyDec) error {
	dists := tr.GasMeter.ReadGasMeter(ctx, tr.ID, tr.Height, burnRatio)

	// Collect reward information for executors and data proxies to emit them
	// in a single event.
	var execRewardAttrs, dataProxyRewardAttrs []sdk.Attribute

	remainingEscrow := tr.GasMeter.GetEscrow()
	for _, dist := range dists {
		var amount math.Int
		switch {
		case dist.Burn != nil:
			amount = math.MinInt(dist.Burn.Amount, remainingEscrow)
			err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, amount)))
			if err != nil {
				return err
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeBurn,
					sdk.NewAttribute(types.AttributeDataRequestID, tr.ID),
					sdk.NewAttribute(types.AttributePostedDataRequestHeight, strconv.FormatUint(tr.Height, 10)),
					sdk.NewAttribute(sdk.AttributeKeyAmount, amount.String()),
				),
			)

		case dist.DataProxyReward != nil:
			amount = math.MinInt(dist.DataProxyReward.Amount, remainingEscrow)
			payoutAddr, err := sdk.AccAddressFromBech32(dist.DataProxyReward.PayoutAddress)
			if err != nil {
				// Should not be reachable because the address has been validated.
				return err
			}
			err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, payoutAddr, sdk.NewCoins(sdk.NewCoin(denom, amount)))
			if err != nil {
				return err
			}

			dataProxyRewardAttrs = append(dataProxyRewardAttrs,
				sdk.NewAttribute(
					types.AttributeDataProxyReward,
					fmt.Sprintf("%s,%s,%s", dist.DataProxyReward.PublicKey, payoutAddr.String(), amount.String()),
				),
			)

		case dist.ExecutorReward != nil:
			amount = math.MinInt(dist.ExecutorReward.Amount, remainingEscrow)
			staker, err := k.GetStaker(ctx, dist.ExecutorReward.Identity)
			if err != nil {
				return err
			}

			// Top-up staked amount to minimum stake and reward the rest.
			reward := amount
			stakeTopup := math.ZeroInt()
			if staker.Staked.LT(minimumStake) {
				stakeTopup = math.MinInt(minimumStake.Sub(staker.Staked), amount)
				staker.Staked = staker.Staked.Add(stakeTopup)
				reward = reward.Sub(stakeTopup)
			}
			staker.PendingWithdrawal = staker.PendingWithdrawal.Add(reward)

			err = k.SetStaker(ctx, staker)
			if err != nil {
				return err
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeExecutorEvent,
					sdk.NewAttribute(types.AttributeExecutor, staker.PublicKey),
					sdk.NewAttribute(types.AttributeTokensStaked, staker.Staked.String()),
					sdk.NewAttribute(types.AttributeTokensPendingWithdrawal, staker.PendingWithdrawal.String()),
					sdk.NewAttribute(types.AttributeMemo, staker.Memo),
				),
			)
			execRewardAttrs = append(execRewardAttrs,
				sdk.NewAttribute(
					types.AttributeExecutorReward,
					fmt.Sprintf("%s,%s,%s", staker.PublicKey, reward.String(), stakeTopup.String()),
				),
			)
		}

		remainingEscrow = remainingEscrow.Sub(amount)
	}

	if len(execRewardAttrs) > 0 {
		execRewardAttrs = append(execRewardAttrs,
			sdk.NewAttribute(types.AttributeDataRequestID, tr.ID),
			sdk.NewAttribute(types.AttributePostedDataRequestHeight, strconv.FormatUint(tr.Height, 10)),
		)
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExecutorRewards, execRewardAttrs...))
	}
	if len(dataProxyRewardAttrs) > 0 {
		dataProxyRewardAttrs = append(dataProxyRewardAttrs,
			sdk.NewAttribute(types.AttributeDataRequestID, tr.ID),
			sdk.NewAttribute(types.AttributePostedDataRequestHeight, strconv.FormatUint(tr.Height, 10)),
		)
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeDataProxyRewards, dataProxyRewardAttrs...))
	}

	// Refund the poster.
	if !remainingEscrow.IsZero() {
		poster, err := sdk.AccAddressFromBech32(tr.GasMeter.GetPoster())
		if err != nil {
			// Should not be reachable because the address has been validated.
			return err
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, poster, sdk.NewCoins(sdk.NewCoin(denom, remainingEscrow)))
		if err != nil {
			return err
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRefund,
				sdk.NewAttribute(types.AttributeDataRequestID, tr.ID),
				sdk.NewAttribute(types.AttributePostedDataRequestHeight, strconv.FormatUint(tr.Height, 10)),
				sdk.NewAttribute(sdk.AttributeKeyAmount, remainingEscrow.String()),
			),
		)
	}

	return nil
}

// MeterProxyGas computes and records the gas consumption of data proxies given
// proxy public keys in basic consensus and the request's replication factor.
func (k Keeper) MeterProxyGas(ctx sdk.Context, gasMeter *types.GasMeter, proxyPubKeys []string, replicationFactor uint64) {
	if len(proxyPubKeys) == 0 || gasMeter.RemainingExecGas() == 0 {
		return
	}

	for _, pubKey := range proxyPubKeys {
		pubKeyBytes, err := hex.DecodeString(pubKey)
		if err != nil {
			k.Logger(ctx).Error("failed to decode proxy public key", "error", err, "public_key", pubKey)
			continue
		}
		proxyConfig, err := k.dataProxyKeeper.GetDataProxyConfig(ctx, pubKeyBytes)
		if err != nil {
			k.Logger(ctx).Error("failed to get proxy config", "error", err, "public_key", pubKey)
			continue
		}

		// Compute the proxy gas used per executor, capping it at the max uint64
		// value and the remaining execution gas.
		// Noting that gasUsed * gasPrice = fee,
		gasUsedPerExecInt := proxyConfig.Fee.Amount.Quo(gasMeter.GasPrice())
		var gasUsedPerExec uint64
		if gasUsedPerExecInt.IsUint64() {
			gasUsedPerExec = min(gasUsedPerExecInt.Uint64(), gasMeter.RemainingExecGas()/replicationFactor)
		} else {
			gasUsedPerExec = min(stdmath.MaxUint64, gasMeter.RemainingExecGas()/replicationFactor)
		}

		gasMeter.ConsumeExecGasForProxy(pubKey, proxyConfig.PayoutAddress, gasUsedPerExec, replicationFactor)
	}
}

// MeterExecutorGasFallback consumes the given fallback gas amount for each
// committer. If any reveal is present, it will only consume gas for committers
// that have also revealed.
func MeterExecutorGasFallback(gasMeter *types.GasMeter, committers, revealers []string, replicationFactor, gasCostFallback uint64) {
	if len(committers) == 0 || gasMeter.RemainingExecGas() == 0 {
		return
	}

	sort.Strings(committers)

	gasLimitPerExec := gasMeter.RemainingExecGas() / replicationFactor

	for _, committer := range committers {
		// If there are reveals, only consume gas for committers that have also
		// revealed.
		if len(revealers) > 0 && !slices.Contains(revealers, committer) {
			continue
		}
		gasUsed := min(gasLimitPerExec, gasCostFallback)
		gasMeter.ConsumeExecGasForExecutor(committer, gasUsed)
	}
}
