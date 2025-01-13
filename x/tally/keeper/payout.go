package keeper

import (
	"fmt"
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

// CalculateUniformPayouts returns payouts for the executors of the given reveals
// and the total gas used for the execution under the uniform reporting scenario.
func CalculateUniformPayouts(reveals []types.RevealBody, execGasLimit uint64, replicationFactor uint16, gasPrice string) ([]types.DistributionMessage, uint64, error) {
	gasUsed := max(reveals[0].GasUsed, execGasLimit/uint64(replicationFactor))
	gasPriceInt, ok := math.NewIntFromString(gasPrice)
	if !ok {
		return nil, 0, fmt.Errorf("invalid gas price: %s", gasPrice) // TODO error
	}
	payout := gasPriceInt.Mul(math.NewIntFromUint64(gasUsed))

	distMsgs := make([]types.DistributionMessage, len(reveals))
	for i, reveal := range reveals {
		distMsgs[i] = types.DistributionMessage{
			Kind: types.DistributionKind{
				ExecutorReward: &types.DistributionExecutorReward{
					Identity: reveal.ID,
					Amount:   payout,
				},
			},
			Type: types.DistributionTypeTimedOut, // TODO check
		}
	}
	return distMsgs, gasUsed, nil
}

// CalculateDivergentPayouts returns payouts for the executors of the given reveals
// and the total gas used for the execution under the divergent reporting scenario.
func CalculateDivergentPayouts(reveals []types.RevealBody, execGasLimit uint64, replicationFactor uint16, gasPrice string) ([]types.DistributionMessage, uint64, error) {
	return nil, 0, nil
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
