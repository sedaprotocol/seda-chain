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

func (k Keeper) GetGasMeterResults(ctx sdk.Context, gasMeter *types.GasMeter, drID string, drHeight int64, burnRatio math.LegacyDec) []types.Distribution {
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeDataRequestID, drID),
		sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatInt(drHeight, 10)),
		sdk.NewAttribute(types.AttributeReducedPayout, strconv.FormatBool(gasMeter.ReducedPayout)),
	}

	// Construct distribution messages to be processed at the end of the function.
	dists := []types.Distribution{}

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

func (k Keeper) ProcessDistributions(ctx sdk.Context, dists []types.Distribution, dr *types.DataRequest, minimumStake math.Int) error {
	// Distribute in order.
	denom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	// TODO Events
	remainingEscrow := dr.Escrow
	for _, dist := range dists {
		if !remainingEscrow.IsPositive() {
			break
		}
		amount := math.ZeroInt()

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
	if remainingEscrow.IsPositive() {
		poster, err := sdk.AccAddressFromBech32(dr.Poster)
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

// MeterExecutorGasUniform computes and records the gas consumption of executors
// when their gas reports are uniformly at "gasReport". If a non-nil outliers
// slice is provided, no gas consumption will be recorded for the executors
// specified as outliers.
func MeterExecutorGasUniform(reveals []types.Reveal, gasReport uint64, outliers []bool, replicationFactor uint64, gasMeter *types.GasMeter) {
	executorGasReport := gasMeter.CorrectExecGasReportWithProxyGas(gasReport)
	gasUsed := min(executorGasReport, gasMeter.RemainingExecGas()/replicationFactor)
	for i, r := range reveals {
		if outliers != nil && outliers[i] {
			continue
		}
		gasMeter.ConsumeExecGasForExecutor(r.Executor, gasUsed)
	}
}

// MeterExecutorGasDivergent computes and records the gas consumption of executors
// when their gas reports are divergent. If a non-nil outliers slice is provided,
// no gas consumption will be recorded for the executors specified as outliers.
func MeterExecutorGasDivergent(reveals []types.Reveal, gasReports []uint64, outliers []bool, replicationFactor uint64, gasMeter *types.GasMeter) {
	var lowestReport uint64
	var lowestReporterIndex int
	adjGasReports := make([]uint64, len(gasReports))
	for i, gasReport := range gasReports {
		executorGasReport := gasMeter.CorrectExecGasReportWithProxyGas(gasReport)
		adjGasReports[i] = min(executorGasReport, gasMeter.RemainingExecGas()/replicationFactor)
		if i == 0 || adjGasReports[i] < lowestReport {
			lowestReporterIndex = i
			lowestReport = adjGasReports[i]
		}
	}
	medianGasUsed := median(adjGasReports)
	totalGasUsed := math.NewIntFromUint64(medianGasUsed*(replicationFactor-1) + min(lowestReport*2, medianGasUsed))
	totalShares := math.NewIntFromUint64(medianGasUsed * (replicationFactor - 1)).Add(math.NewIntFromUint64(lowestReport * 2))
	var lowestGasUsed, regGasUsed uint64
	if totalShares.IsZero() {
		lowestGasUsed = 0
		regGasUsed = 0
	} else {
		lowestGasUsed = math.NewIntFromUint64(lowestReport * 2).Mul(totalGasUsed).Quo(totalShares).Uint64()
		regGasUsed = math.NewIntFromUint64(medianGasUsed).Mul(totalGasUsed).Quo(totalShares).Uint64()
	}
	for i, r := range reveals {
		if outliers != nil && outliers[i] {
			continue
		}
		gasUsed := regGasUsed
		if i == lowestReporterIndex {
			gasUsed = lowestGasUsed
		}
		gasMeter.ConsumeExecGasForExecutor(r.Executor, gasUsed)
	}
}

func median(arr []uint64) uint64 {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	n := len(arr)
	if n%2 == 0 {
		return (arr[n/2-1] + arr[n/2]) / 2
	}
	return arr[n/2]
}

// areGasReportsUniform returns true if the gas reports of the given reveals are
// uniform.
func areGasReportsUniform(reports []uint64) bool {
	if len(reports) <= 1 {
		return true
	}
	firstGas := reports[0]
	for i := 1; i < len(reports); i++ {
		if reports[i] != firstGas {
			return false
		}
	}
	return true
}
