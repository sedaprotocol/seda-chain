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

// GasCalculation stores the results of the canonical gas consumption calculations.
type GasCalculation struct {
	Burn          uint64 // filter and tally gas to be burned
	Proxies       []ProxyGasUsed
	Executors     []ExecutorGasUsed
	ReducedPayout bool
}

type ProxyGasUsed struct {
	PublicKey []byte
	Amount    math.Int
}

type ExecutorGasUsed struct {
	PublicKey string
	Amount    math.Int
}

// DistributionsFromGasCalculation constructs a list of distribution messages
// to be sent to the core contract based on the given gas calculation. It takes
// the request ID and height for event emission.
func (k Keeper) DistributionsFromGasCalculation(ctx sdk.Context, reqID string, reqHeight uint64, gasCalc GasCalculation, gasPrice math.Int, burnRatio math.LegacyDec) []types.Distribution {
	dists := []types.Distribution{}
	attrs := []sdk.Attribute{
		sdk.NewAttribute(types.AttributeDataRequestID, reqID),
		sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatUint(reqHeight, 10)),
		sdk.NewAttribute(types.AttributeReducedPayout, strconv.FormatBool(gasCalc.ReducedPayout)),
	}

	// First distribution message is the combined burn.
	burn := types.NewBurn(math.NewIntFromUint64(gasCalc.Burn), gasPrice)
	dists = append(dists, burn)
	attrs = append(attrs, sdk.NewAttribute(types.AttributeTallyGas, strconv.FormatUint(gasCalc.Burn, 10)))

	// Append distribution messages for data proxies.
	for _, proxy := range gasCalc.Proxies {
		proxyDist := types.NewDataProxyReward(proxy.PublicKey, proxy.Amount, gasPrice)
		dists = append(dists, proxyDist)
		attrs = append(attrs, sdk.NewAttribute(types.AttributeDataProxyGas,
			fmt.Sprintf("%s,%s", hex.EncodeToString(proxy.PublicKey), proxy.Amount.String())))
	}

	// Append distribution messages for executors, burning a portion of their
	// payouts in case of a reduced payout scenario.
	reducedPayoutBurn := math.ZeroInt()
	for _, executor := range gasCalc.Executors {
		payoutAmt := executor.Amount
		if gasCalc.ReducedPayout {
			burnAmt := burnRatio.MulInt(executor.Amount).TruncateInt()
			payoutAmt = executor.Amount.Sub(burnAmt)
			reducedPayoutBurn = reducedPayoutBurn.Add(burnAmt.Mul(gasPrice))
		}

		executorDist := types.NewExecutorReward(executor.PublicKey, payoutAmt, gasPrice)
		dists = append(dists, executorDist)
		attrs = append(attrs, sdk.NewAttribute(types.AttributeExecutorGas,
			fmt.Sprintf("%s,%s", executor.PublicKey, payoutAmt.String())))
	}

	dists[0].Burn.Amount = dists[0].Burn.Amount.Add(reducedPayoutBurn.Mul(gasPrice))
	attrs = append(attrs, sdk.NewAttribute(types.AttributeReducedPayoutBurn, reducedPayoutBurn.String()))

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeGasCalculation, attrs...))

	return dists
}

// ProxyGasUsed returns the results of data proxy gas calculations, the total
// gas used, and the remaining gas limit per executor.
func (k Keeper) ProxyGasUsed(ctx sdk.Context, proxyPubKeys []string, gasPrice math.Int, gasLimit uint64, replicationFactor uint16) ([]ProxyGasUsed, uint64) {
	if len(proxyPubKeys) == 0 || gasLimit == 0 {
		return nil, 0
	}
	gasLimitPerExec := gasLimit / uint64(replicationFactor)

	var totalGasUsed uint64
	calcs := []ProxyGasUsed{}
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

		gasUsedPerExec := proxyConfig.Fee.Amount.Quo(gasPrice).Uint64()
		gasUsedPerExec = min(gasUsedPerExec, gasLimitPerExec)
		gasUsed := gasUsedPerExec * uint64(replicationFactor)

		calcs = append(calcs, ProxyGasUsed{
			PublicKey: pubKeyBytes,
			Amount:    math.NewIntFromUint64(gasUsed),
		})

		totalGasUsed += gasUsed
		gasLimitPerExec -= gasUsedPerExec
	}
	return calcs, totalGasUsed
}

// CommitGasUsed processes gas consumption of committers of a given data request.
func CommitGasUsed(req types.Request, gasCostCommit, gasLimit uint64) ([]ExecutorGasUsed, uint64) {
	if len(req.Commits) == 0 || gasLimit == 0 {
		return nil, 0
	}
	gasLimitPerExec := gasLimit / uint64(req.ReplicationFactor)

	i := 0
	committers := make([]string, len(req.Commits))
	for k := range req.Commits {
		committers[i] = k
		i++
	}
	sort.Strings(committers)

	var totalGasUsed uint64
	calcs := make([]ExecutorGasUsed, len(committers))
	for i, committer := range committers {
		gasUsed := min(gasLimitPerExec, gasCostCommit)
		calcs[i] = ExecutorGasUsed{
			PublicKey: committer,
			Amount:    math.NewIntFromUint64(gasUsed),
		}
		totalGasUsed += gasUsed
	}
	return calcs, totalGasUsed
}

// ProcessGasUsedUniform calculates the canonical gas consumption by the
// executors when their gas reports are uniformly at "gasReport". It also returns
// the total execution gas consumption.
func ExecutorGasUsedUniform(executors []string, gasReport, gasLimit uint64, replicationFactor uint16) ([]ExecutorGasUsed, uint64) {
	gasUsed := min(gasReport, gasLimit/uint64(replicationFactor))

	var totalGasUsed uint64
	calcs := make([]ExecutorGasUsed, len(executors))
	for i, executor := range executors {
		calcs[i] = ExecutorGasUsed{
			PublicKey: executor,
			Amount:    math.NewIntFromUint64(gasUsed),
		}
		totalGasUsed += gasUsed
	}
	return calcs, totalGasUsed
}

// ProcessGasUsedDivergent calculates the canonical gas consumption by the
// executors given their divergent gas reports. It also returns the total
// execution gas consumption.
func ExecutorGasUsedDivergent(executors []string, gasReports []uint64, gasLimit uint64, replicationFactor uint16) ([]ExecutorGasUsed, uint64) {
	var lowestReport uint64
	var lowestReporterIndex int
	adjGasReports := make([]uint64, len(gasReports))
	for i, gasReport := range gasReports {
		adjGasReports[i] = min(gasReport, gasLimit/uint64(replicationFactor))
		if i == 0 || adjGasReports[i] < lowestReport {
			lowestReporterIndex = i
			lowestReport = adjGasReports[i]
		}
	}
	medianGasUsed := median(adjGasReports)
	totalGasUsed := medianGasUsed*uint64(replicationFactor-1) + min(lowestReport*2, medianGasUsed)
	totalShares := medianGasUsed*uint64(replicationFactor-1) + lowestReport*2
	lowestGasUsed := math.NewIntFromUint64(lowestReport * 2 * totalGasUsed / totalShares)
	regGasUsed := math.NewIntFromUint64(medianGasUsed * totalGasUsed / totalShares)

	calcs := make([]ExecutorGasUsed, len(executors))
	for i, executor := range executors {
		gasUsed := regGasUsed
		if i == lowestReporterIndex {
			gasUsed = lowestGasUsed
		}
		calcs[i] = ExecutorGasUsed{
			PublicKey: executor,
			Amount:    gasUsed,
		}
	}
	return calcs, totalGasUsed
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
