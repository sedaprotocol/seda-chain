package keeper

import (
	"encoding/hex"
	stdmath "math"
	"slices"
	"sort"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// ChargeGasCosts charges gas costs to the escrow funds based on gas meter
// reading. Remaining escrow funds are refunded to the data request poster.
func (k Keeper) ChargeGasCosts(ctx sdk.Context, denom string, tr *TallyResult, minimumStake math.Int, burnRatio math.LegacyDec) error {
	dists := tr.GasMeter.ReadGasMeter(ctx, tr.ID, tr.Height, burnRatio)

	remainingEscrow := tr.GasMeter.GetEscrow()

	// TODO Events
	var amount math.Int
	for _, dist := range dists {
		switch {
		case dist.Burn != nil:
			amount = math.MinInt(dist.Burn.Amount, remainingEscrow)
			err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, amount)))
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
func (k Keeper) MeterProxyGas(ctx sdk.Context, gasMeter *types.GasMeter, proxyPubKeys []string, replicationFactor uint64) {
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

// MeterExecutorGasFallback consumes the given fallback gas amount for each
// committer. If any reveal is present, it will only consume gas for committers
// that have also revealed.
func MeterExecutorGasFallback(gasMeter *types.GasMeter, committers, revealers []string, replicationFactor, gasCostFallback uint64) {
	if len(committers) == 0 || gasMeter.RemainingExecGas() == 0 {
		return
	}

	sort.Strings(committers)

	gasLimitPerExec := gasMeter.RemainingExecGas() / replicationFactor

	for _, committer := range committers {
		// If there are reveals, only consume gas for committers that have also
		// revealed.
		if len(revealers) > 0 && !slices.Contains(revealers, committer) {
			continue
		}
		gasUsed := min(gasLimitPerExec, gasCostFallback)
		gasMeter.ConsumeExecGasForExecutor(committer, gasUsed)
	}
}
