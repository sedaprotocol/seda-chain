package keeper

import (
	"encoding/hex"
	"fmt"
	stdmath "math"
	"sort"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// ReadGasMeter reads the given gas meter to construct a list of distributions.
func (k Keeper) ReadGasMeter(ctx sdk.Context, gasMeter *types.GasMeter, drID string, drHeight uint64, burnRatio math.LegacyDec) []types.Distribution {
	dists := []types.Distribution{}
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeDataRequestID, drID),
		sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatUint(drHeight, 10)),
		sdk.NewAttribute(types.AttributeReducedPayout, strconv.FormatBool(gasMeter.ReducedPayout)),
	}

	// First distribution message is the combined burn.
	burn := types.NewBurn(math.NewIntFromUint64(gasMeter.TallyGasUsed()), gasMeter.GasPrice())
	dists = append(dists, burn)
	attrs = append(attrs, sdk.NewAttribute(types.AttributeTallyGas, strconv.FormatUint(gasMeter.TallyGasUsed(), 10)))

	// Append distribution messages for data proxies.
	for _, proxy := range gasMeter.GetProxyGasUsed(drID, ctx.BlockHeight()) {
		proxyDist := types.NewDataProxyReward(proxy.PublicKey, proxy.PayoutAddress, proxy.Amount, gasMeter.GasPrice())
		dists = append(dists, proxyDist)
		attrs = append(attrs, sdk.NewAttribute(types.AttributeDataProxyGas,
			fmt.Sprintf("%s,%s,%s", proxy.PublicKey, proxy.PayoutAddress, proxy.Amount.String())))
	}

	// Append distribution messages for executors, burning a portion of their
	// payouts in case of a reduced payout scenario.
	reducedPayoutBurn := math.ZeroInt()
	for _, executor := range gasMeter.GetExecutorGasUsed() {
		payoutAmt := executor.Amount
		if gasMeter.ReducedPayout {
			burnAmt := burnRatio.MulInt(executor.Amount).TruncateInt()
			payoutAmt = executor.Amount.Sub(burnAmt)
			reducedPayoutBurn = reducedPayoutBurn.Add(burnAmt)
		}

		executorDist := types.NewExecutorReward(executor.PublicKey, payoutAmt, gasMeter.GasPrice())
		dists = append(dists, executorDist)
		attrs = append(attrs, sdk.NewAttribute(types.AttributeExecutorGas,
			fmt.Sprintf("%s,%s", executor.PublicKey, payoutAmt.String())))
	}

	dists[0].Burn.Amount = dists[0].Burn.Amount.Add(reducedPayoutBurn.Mul(gasMeter.GasPrice()))
	attrs = append(attrs, sdk.NewAttribute(types.AttributeReducedPayoutBurn, reducedPayoutBurn.String()))

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeGasMeter, attrs...))

	return dists
}

// ChargeGasCosts charges gas costs to the escrow funds based on gas meter
// reading. Remaining escrow funds are refunded to the data request poster.
func (k Keeper) ChargeGasCosts(ctx sdk.Context, tr *TallyResult, minimumStake math.Int, burnRatio math.LegacyDec) error {
	dists := k.ReadGasMeter(ctx, tr.GasMeter, tr.ID, tr.Height, burnRatio)

	// Distribute in order.
	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	remainingEscrow := tr.GasMeter.GetEscrow()

	// TODO Events
	var amount math.Int
	for _, dist := range dists {
		switch {
		case dist.Burn != nil:
			amount = math.MinInt(dist.Burn.Amount, remainingEscrow)
			err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, amount)))
			if err != nil {
				return err
			}

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

		case dist.ExecutorReward != nil:
			amount = math.MinInt(dist.ExecutorReward.Amount, remainingEscrow)
			staker, err := k.GetStaker(ctx, dist.ExecutorReward.Identity)
			if err != nil {
				return err
			}

			// Top up staked amount to minimum stake.
			topup := math.ZeroInt()
			if staker.Staked.LT(minimumStake) {
				topup = math.MinInt(minimumStake.Sub(staker.Staked), amount)
				staker.Staked = staker.Staked.Add(topup)
				remainingEscrow = remainingEscrow.Sub(topup)
			}
			staker.PendingWithdrawal = staker.PendingWithdrawal.Add(amount.Sub(topup))

			err = k.SetStaker(ctx, staker)
			if err != nil {
				return err
			}
		}

		remainingEscrow = remainingEscrow.Sub(amount)
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
	}

	return nil
}

// MeterProxyGas computes and records the gas consumption of data proxies given
// proxy public keys in basic consensus and the request's replication factor.
func (k Keeper) MeterProxyGas(ctx sdk.Context, proxyPubKeys []string, replicationFactor uint64, gasMeter *types.GasMeter) {
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

// MeterExecutorGasFallback computes and records the gas consumption of committers
// of a data request when basic consensus has not been reached. If checkReveal is
// set to true, it will only consume gas for committers that have also revealed.
func MeterExecutorGasFallback(req types.DataRequest, gasCostFallback uint64, gasMeter *types.GasMeter) {
	if len(req.Commits) == 0 || gasMeter.RemainingExecGas() == 0 {
		return
	}

	i := 0
	committers := make([]string, len(req.Commits))
	for k := range req.Commits {
		committers[i] = k
		i++
	}
	sort.Strings(committers)

	gasLimitPerExec := gasMeter.RemainingExecGas() / uint64(req.ReplicationFactor)

	hasReveals := len(req.Reveals) > 0
	for _, committer := range committers {
		// If there are reveals, only consume gas for committers that have also
		// revealed.
		if hasReveals {
			if _, ok := req.Reveals[committer]; !ok {
				continue
			}
		}
		gasUsed := min(gasLimitPerExec, gasCostFallback)
		gasMeter.ConsumeExecGasForExecutor(committer, gasUsed)
	}
}
