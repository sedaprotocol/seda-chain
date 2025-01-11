package keeper

import (
	"sort"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

// CalculateCommitterPayouts constructs distribution messages that
// pay the fixed gas cost for each commiter of a given data request.
func (k Keeper) CalculateCommitterPayouts(ctx sdk.Context, req types.Request) (types.DistributionMessages, error) {
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
					Amount:   math.NewIntFromUint64(gasCost),
				},
			},
			Type: types.DistributionTypeTimedOut,
		}
	}
	result.Messages = distMsgs
	return result, nil
}

// TODO: This will become more complex when we introduce incentives.
func calculateExecGasUsed(reveals []types.RevealBody) uint64 {
	var execGasUsed uint64
	for _, reveal := range reveals {
		execGasUsed += reveal.GasUsed
	}
	return execGasUsed
}
