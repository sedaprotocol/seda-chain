package keeper

import (
	"encoding/hex"
	"fmt"
	stdmath "math"
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
	proxyReports := gasMeter.GetProxyGasUsed(reqID, ctx.BlockHeight())
	executorReports := gasMeter.GetExecutorGasUsed()

	dists := make([]types.Distribution, 0, 1+len(proxyReports)+len(executorReports))
	attrs := make([]sdk.Attribute, 0, 4+len(proxyReports)+len(executorReports))

	// First distribution message is the combined burn.
	burn := types.NewBurn(math.NewIntFromUint64(gasMeter.TallyGasUsed()), gasMeter.GasPrice())
	dists = append(dists, burn)
	attrs = append(attrs, []sdk.Attribute{
		sdk.NewAttribute(types.AttributeDataRequestID, reqID),
		sdk.NewAttribute(types.AttributeDataRequestHeight, strconv.FormatUint(reqHeight, 10)),
		sdk.NewAttribute(types.AttributeReducedPayout, strconv.FormatBool(gasMeter.ReducedPayout)),
		sdk.NewAttribute(types.AttributeTallyGas, strconv.FormatUint(gasMeter.TallyGasUsed(), 10)),
	}...)

	// Append distribution messages for data proxies.
	for _, proxy := range proxyReports {
		proxyDist := types.NewDataProxyReward(proxy.PublicKey, proxy.PayoutAddress, proxy.Amount, gasMeter.GasPrice())
		dists = append(dists, proxyDist)
		attrs = append(attrs, sdk.NewAttribute(types.AttributeDataProxyGas,
			fmt.Sprintf("%s,%s,%s", proxy.PublicKey, proxy.PayoutAddress, proxy.Amount.String())))
	}

	// Append distribution messages for executors, burning a portion of their
	// payouts in case of a reduced payout scenario.
	reducedPayoutBurn := math.ZeroInt()
	for _, executor := range executorReports {
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

	// Add the reduced payout burn (may be zero) to the burn distribution.
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

		// Compute the proxy gas used per executor, capping it at the max uint64
		// value and the remaining execution gas.
		// Noting that gasUsed * gasPrice = fee,
		gasUsedPerExecInt := proxyConfig.Fee.Amount.Quo(gasMeter.GasPrice())
		var gasUsedPerExec uint64
		if gasUsedPerExecInt.IsUint64() {
			gasUsedPerExec = min(gasUsedPerExecInt.Uint64(), gasMeter.RemainingExecGas()/uint64(replicationFactor))
		} else {
			gasUsedPerExec = min(stdmath.MaxUint64, gasMeter.RemainingExecGas()/uint64(replicationFactor))
		}

		gasMeter.ConsumeExecGasForProxy(pubKey, proxyConfig.PayoutAddress, gasUsedPerExec, replicationFactor)
	}
}

// MeterExecutorGasFallback computes and records the gas consumption of committers
// of a data request when basic consensus has not been reached. If checkReveal is
// set to true, it will only consume gas for committers that have also revealed.
func MeterExecutorGasFallback(req types.Request, gasCostFallback uint64, gasMeter *types.GasMeter) {
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
