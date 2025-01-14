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
func (k Keeper) CalculateDataProxyPayouts(ctx sdk.Context, proxyPubKeys []string, gasPrice math.Int) (types.DistributionMessages, uint64, error) {
	var result types.DistributionMessages
	if len(proxyPubKeys) == 0 {
		return result, 0, nil
	}

	gasUsed := math.NewInt(0)
	distMsgs := make([]types.DistributionMessage, len(proxyPubKeys))
	for i, pubKey := range proxyPubKeys {
		pubKeyBytes, err := hex.DecodeString(pubKey)
		if err != nil {
			return types.DistributionMessages{}, 0, err
		}
		proxyConfig, err := k.dataProxyKeeper.GetDataProxyConfig(ctx, pubKeyBytes)
		if err != nil {
			return types.DistributionMessages{}, 0, err
		}
		gasUsed = gasUsed.Add(proxyConfig.Fee.Amount.Quo(gasPrice))

		distMsgs[i] = types.DistributionMessage{
			Kind: types.DistributionKind{
				ExecutorReward: &types.DistributionExecutorReward{
					Identity: pubKey,
					Amount:   proxyConfig.Fee.Amount,
				},
			},
		}
	}
	result.Messages = distMsgs
	return types.DistributionMessages{}, gasUsed.Uint64(), nil // TODO may panic
}

// CalculateCommitterPayouts returns the fixed payouts for the committers of a
// given data request.
func (k Keeper) CalculateCommitterPayouts(ctx sdk.Context, req types.Request, gasPrice math.Int) (types.DistributionMessages, error) {
	var result types.DistributionMessages
	if len(req.Commits) == 0 {
		return result, nil
	}

	gasCost, err := k.GetGasCostCommitment(ctx)
	if err != nil {
		return types.DistributionMessages{}, err
	}
	payout := gasPrice.Mul(math.NewIntFromUint64(gasCost))

	i := 0
	committers := make([]string, len(req.Commits))
	for k := range req.Commits {
		committers[i] = k
		i++
	}
	sort.Strings(committers)

	distMsgs := make([]types.DistributionMessage, len(committers))
	for i, committer := range committers {
		distMsgs[i] = types.DistributionMessage{
			Kind: types.DistributionKind{
				ExecutorReward: &types.DistributionExecutorReward{
					Identity: committer,
					Amount:   payout,
				},
			},
		}
	}
	result.Messages = distMsgs
	return result, nil
}

// CalculateUniformPayouts calculates payouts for the executors when their gas
// reports are uniformly at "gasReport". It also returns the total execution gas
// consumption.
func CalculateUniformPayouts(executors []string, gasReport, execGasLimit uint64, replicationFactor uint16, gasPrice math.Int) ([]types.DistributionMessage, uint64) {
	adjGasUsed := max(gasReport, execGasLimit/uint64(replicationFactor))
	payout := gasPrice.Mul(math.NewIntFromUint64(adjGasUsed))

	distMsgs := make([]types.DistributionMessage, len(executors))
	for i := range executors {
		distMsgs[i] = types.DistributionMessage{
			Kind: types.DistributionKind{
				ExecutorReward: &types.DistributionExecutorReward{
					Identity: executors[i],
					Amount:   payout,
				},
			},
		}
	}
	return distMsgs, adjGasUsed * uint64(replicationFactor)
}

// CalculateDivergentPayouts calculates payouts for the executors given their
// divergent gas reports. It also returns the total execution gas consumption.
// It assumes that the i-th executor is the one who revealed the i-th reveal.
func CalculateDivergentPayouts(executors []string, gasReports []uint64, execGasLimit uint64, replicationFactor uint16, gasPrice math.Int) ([]types.DistributionMessage, uint64) {
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

	distMsgs := make([]types.DistributionMessage, len(executors))
	for i, executor := range executors {
		payout := normalPayout
		if i == lowestReporterIndex {
			payout = lowestPayout
		}
		distMsgs[i] = types.DistributionMessage{
			Kind: types.DistributionKind{
				ExecutorReward: &types.DistributionExecutorReward{
					Identity: executor,
					Amount:   payout,
				},
			},
		}
	}
	return distMsgs, totalGasUsed
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
