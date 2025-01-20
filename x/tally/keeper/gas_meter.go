package keeper

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

// DistributionsFromGasMeter constructs a list of distribution messages to be
// sent to the core contract based on the given gas meter. It takes the ID and
// the height of the request for event emission.
func (k Keeper) DistributionsFromGasMeter(ctx sdk.Context, reqID string, reqHeight uint64, gasMeter *types.GasMeter, burnRatio math.LegacyDec) []types.Distribution {
	dists := []types.Distribution{}
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeDataRequestID, reqID),
		sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatUint(reqHeight, 10)),
		sdk.NewAttribute(types.AttributeReducedPayout, strconv.FormatBool(gasMeter.ReducedPayout)),
	}

	// First distribution message is the combined burn.
	burn := types.NewBurn(math.NewIntFromUint64(gasMeter.Burn), gasMeter.GasPrice())
	dists = append(dists, burn)
	attrs = append(attrs, sdk.NewAttribute(types.AttributeTallyGas, strconv.FormatUint(gasMeter.Burn, 10)))

	// Append distribution messages for data proxies.
	for _, proxy := range gasMeter.Proxies {
		proxyDist := types.NewDataProxyReward(proxy.PublicKey, proxy.Amount, gasMeter.GasPrice())
		dists = append(dists, proxyDist)
		attrs = append(attrs, sdk.NewAttribute(types.AttributeDataProxyGas,
			fmt.Sprintf("%s,%s", hex.EncodeToString(proxy.PublicKey), proxy.Amount.String())))
	}

	// Append distribution messages for executors, burning a portion of their
	// payouts in case of a reduced payout scenario.
	reducedPayoutBurn := math.ZeroInt()
	for _, executor := range gasMeter.Executors {
		payoutAmt := executor.Amount
		if gasMeter.ReducedPayout {
			burnAmt := burnRatio.MulInt(executor.Amount).TruncateInt()
			payoutAmt = executor.Amount.Sub(burnAmt)
			reducedPayoutBurn = reducedPayoutBurn.Add(burnAmt.Mul(gasMeter.GasPrice()))
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

// MeterProxyGas computes and records the gas consumption of data proxies given
// proxy public keys in basic consensus and the request's replication factor.
func (k Keeper) MeterProxyGas(ctx sdk.Context, proxyPubKeys []string, replicationFactor uint16, gasMeter *types.GasMeter) {
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

		gasUsedPerExec := proxyConfig.Fee.Amount.Quo(gasMeter.GasPrice()).Uint64()
		gasUsedPerExec = min(gasUsedPerExec, gasMeter.RemainingExecGas()/uint64(replicationFactor))
		gasUsed := gasUsedPerExec * uint64(replicationFactor)

		gasMeter.ConsumeExecGasForProxy(pubKeyBytes, gasUsed)
	}
}

// MeterExecutorGasFallback computes and records the gas consumption of committers
// of a data request when basic consensus has not been reached.
func MeterExecutorGasFallback(req types.Request, gasCostCommit uint64, gasMeter *types.GasMeter) {
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
	for _, committer := range committers {
		gasUsed := min(gasLimitPerExec, gasCostCommit)
		gasMeter.ConsumeExecGasForExecutor(committer, gasUsed)
	}
}

// MeterExecutorGasUniform computes and records the gas consumption of executors
// when their gas reports are uniformly at "gasReport".
func MeterExecutorGasUniform(executors []string, gasReport uint64, replicationFactor uint16, gasMeter *types.GasMeter) {
	gasUsed := min(gasReport, gasMeter.RemainingExecGas()/uint64(replicationFactor))
	for _, executor := range executors {
		gasMeter.ConsumeExecGasForExecutor(executor, gasUsed)
	}
}

// MeterExecutorGasDivergent computes and records the gas consumption of executors
// when their gas reports are divergent.
func MeterExecutorGasDivergent(executors []string, gasReports []uint64, replicationFactor uint16, gasMeter *types.GasMeter) {
	var lowestReport uint64
	var lowestReporterIndex int
	adjGasReports := make([]uint64, len(gasReports))
	for i, gasReport := range gasReports {
		adjGasReports[i] = min(gasReport, gasMeter.RemainingExecGas()/uint64(replicationFactor))
		if i == 0 || adjGasReports[i] < lowestReport {
			lowestReporterIndex = i
			lowestReport = adjGasReports[i]
		}
	}
	medianGasUsed := median(adjGasReports)
	totalGasUsed := medianGasUsed*uint64(replicationFactor-1) + min(lowestReport*2, medianGasUsed)
	totalShares := medianGasUsed*uint64(replicationFactor-1) + lowestReport*2
	lowestGasUsed := lowestReport * 2 * totalGasUsed / totalShares
	regGasUsed := medianGasUsed * totalGasUsed / totalShares

	for i, executor := range executors {
		gasUsed := regGasUsed
		if i == lowestReporterIndex {
			gasUsed = lowestGasUsed
		}
		gasMeter.ConsumeExecGasForExecutor(executor, gasUsed)
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
