package keeper

import (
	"encoding/hex"
	"sort"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

// CalculateDataProxyPayouts returns payouts for the data proxies and
// returns the the gas used by the data proxies per executor.
func (k Keeper) CalculateDataProxyPayouts(ctx sdk.Context, proxyPubKeys []string, gasPrice math.Int) ([]types.Distribution, uint64) {
	if len(proxyPubKeys) == 0 {
		return nil, 0
	}
	gasUsed := math.ZeroInt()
	dists := make([]types.Distribution, len(proxyPubKeys))
	for i, pubKey := range proxyPubKeys {
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

		gasUsed = gasUsed.Add(proxyConfig.Fee.Amount.Quo(gasPrice))
		dists[i] = types.NewDataProxyReward(pubKeyBytes, proxyConfig.Fee.Amount)

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeDataProxyReward,
			sdk.NewAttribute(types.AttributeProxyPubKey, pubKey),
			sdk.NewAttribute(sdk.AttributeKeyAmount, proxyConfig.Fee.Amount.String()),
		))
	}
	return dists, gasUsed.Uint64()
}

// CalculateCommitterPayouts returns distribution messages of a given payout
// amount to the committers of a data request. The messages are sorted by
// committer public key.
func (k Keeper) CalculateCommitterPayouts(ctx sdk.Context, req types.Request, gasCostCommit uint64, gasPrice math.Int) []types.Distribution {
	if len(req.Commits) == 0 {
		return nil
	}

	i := 0
	committers := make([]string, len(req.Commits))
	for k := range req.Commits {
		committers[i] = k
		i++
	}
	sort.Strings(committers)

	payout := gasPrice.Mul(math.NewIntFromUint64(gasCostCommit))
	dists := make([]types.Distribution, len(committers))
	for i, committer := range committers {
		dists[i] = types.NewExecutorReward(committer, payout)

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExecutorRewardCommit,
			sdk.NewAttribute(types.AttributeExecutor, committer),
			sdk.NewAttribute(sdk.AttributeKeyAmount, payout.String()),
		))
	}
	return dists
}

// CalculateUniformPayouts calculates payouts for the executors when their gas
// reports are uniformly at "gasReport". It also returns the total execution gas
// consumption and, in case of reduced payout, the amount to be burned.
func (k Keeper) CalculateUniformPayouts(ctx sdk.Context, executors []string, gasReport, execGasLimit uint64, replicationFactor uint16, gasPrice math.Int, reducedPayout bool) ([]types.Distribution, uint64, math.Int) {
	adjGasUsed := min(gasReport, execGasLimit/uint64(replicationFactor))
	payoutAmt := gasPrice.Mul(math.NewIntFromUint64(adjGasUsed))
	burnAmt := math.ZeroInt()
	if reducedPayout {
		burnRatio := math.LegacyNewDecWithPrec(2, 1) // burn 20%
		burnAmt := burnRatio.MulInt(payoutAmt).TruncateInt()
		payoutAmt = payoutAmt.Sub(burnAmt)
	}

	totalBurn := math.ZeroInt()
	dists := make([]types.Distribution, len(executors))
	for i, executor := range executors {
		if burnAmt.IsPositive() {
			totalBurn = totalBurn.Add(burnAmt)
		}
		dists[i] = types.NewExecutorReward(executor, payoutAmt)

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExecutorRewardUniform,
			sdk.NewAttribute(types.AttributeExecutor, executor),
			sdk.NewAttribute(sdk.AttributeKeyAmount, payoutAmt.String()),
		))
	}
	return dists, adjGasUsed * uint64(replicationFactor), totalBurn
}

// CalculateDivergentPayouts calculates payouts for the executors given their
// divergent gas reports. It also returns the total execution gas consumption.
// It assumes that the i-th executor is the one who revealed the i-th reveal.
func (k Keeper) CalculateDivergentPayouts(ctx sdk.Context, executors []string, gasReports []uint64, execGasLimit uint64, replicationFactor uint16, gasPrice math.Int, reducedPayout bool) ([]types.Distribution, uint64, math.Int) {
	adjGasUsed := make([]uint64, len(gasReports))
	var lowestGasUsed uint64
	var lowestReporterIndex int
	for i, gasReport := range gasReports {
		adjGasUsed[i] = min(gasReport, execGasLimit/uint64(replicationFactor))
		if i == 0 || adjGasUsed[i] < lowestGasUsed {
			lowestReporterIndex = i
			lowestGasUsed = adjGasUsed[i]
		}
	}
	medianGasUsed := median(adjGasUsed)
	totalGasUsed := medianGasUsed*uint64(replicationFactor-1) + min(lowestGasUsed*2, medianGasUsed)
	totalShares := medianGasUsed*uint64(replicationFactor-1) + lowestGasUsed*2
	lowestPayout := gasPrice.Mul(math.NewIntFromUint64(lowestGasUsed * 2 * totalGasUsed / totalShares))
	normalPayout := gasPrice.Mul(math.NewIntFromUint64(medianGasUsed * totalGasUsed / totalShares))

	burnRatio := math.LegacyNewDecWithPrec(2, 1) // burn 20% in case of reduced payout

	totalBurn := math.ZeroInt()
	dists := make([]types.Distribution, len(executors))
	for i, executor := range executors {
		payout := normalPayout
		if i == lowestReporterIndex {
			payout = lowestPayout
		}
		if reducedPayout {
			burnAmt := burnRatio.MulInt(payout).TruncateInt()
			totalBurn = totalBurn.Add(burnAmt)
			payout = payout.Sub(burnAmt)
		}
		dists[i] = types.NewExecutorReward(executor, payout)

		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeExecutorRewardUniform,
			sdk.NewAttribute(types.AttributeExecutor, executor),
			sdk.NewAttribute(sdk.AttributeKeyAmount, payout.String()),
		))
	}
	return dists, totalGasUsed, totalBurn
}

func median(arr []uint64) uint64 {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	return arr[len(arr)/2]
}

// areGasReportsUniform returns true if the gas reports of the given reveals are
// uniform.
func areGasReportsUniform(reports []uint64) bool {
	if len(reports) == 0 {
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
