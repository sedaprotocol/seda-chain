package keeper

import (
	"sort"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

// CalculateCommitterPayouts constructs distribution messages that
// pay the fixed gas cost for each commiter of a given data request.
func (k Keeper) CalculateCommitterPayouts(ctx sdk.Context, req types.Request, gasPrice math.Int) (types.DistributionMessages, error) {
	result := types.DistributionMessages{
		Messages:   []types.DistributionMessage{},
		RefundType: types.DistributionTypeTimedOut,
	}
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
			Type: types.DistributionTypeTimedOut,
		}
	}
	result.Messages = distMsgs
	return result, nil
}

// CalculateUniformPayouts calculates payouts for the executors when their gas
// reports are uniformly at "gasUsed". It also returns the total execution gas
// consumption.
func CalculateUniformPayouts(executors []string, gasUsed, execGasLimit uint64, replicationFactor uint16, gasPrice math.Int) ([]types.DistributionMessage, uint64) {
	adjGasUsed := max(gasUsed, execGasLimit/uint64(replicationFactor))
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
			Type: types.DistributionTypeExecutorReward,
		}
	}
	return distMsgs, adjGasUsed * uint64(replicationFactor)
}

// CalculateDivergentPayouts calculates payouts for the executors of the given
// reveals when their gas reports are divergent. It also returns the total
// execution gas consumption.
// It assumes that the i-th executor is the one who revealed the i-th reveal.
func CalculateDivergentPayouts(executors []string, reveals []types.RevealBody, execGasLimit uint64, replicationFactor uint16, gasPrice math.Int) ([]types.DistributionMessage, uint64) {
	adjGasUsed := make([]uint64, len(reveals))
	var lowestGasUsed uint64
	var lowestReporterIndex int
	for i, reveal := range reveals {
		adjGasUsed[i] = min(reveal.GasUsed, execGasLimit/uint64(replicationFactor))
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
			Type: types.DistributionTypeExecutorReward,
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
func areGasReportsUniform(reveals []types.RevealBody) bool {
	if len(reveals) == 0 {
		return true
	}
	firstGas := reveals[0].GasUsed
	for i := 1; i < len(reveals); i++ {
		if reveals[i].GasUsed != firstGas {
			return false
		}
	}
	return true
}
