package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/keeper"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func TestMeterExecutorGasFallback(t *testing.T) {
	ctx := sdktestutil.DefaultContext(
		storetypes.NewKVStoreKey(types.StoreKey),
		storetypes.NewTransientStoreKey("transient_test"),
	)

	tests := []struct {
		name                    string
		dataRequest             *types.DataRequest
		tallyGasLimit           uint64
		baseGasCost             uint64
		fallbackGasCost         uint64
		expectedExecutorRewards []*types.DistributionExecutorReward
	}{
		{
			name: "reward all committers",
			dataRequest: &types.DataRequest{
				Commits: map[string][]byte{
					"executorA": []byte("commit"),
					"executorB": []byte("commit"),
				},
				Reveals: map[string]bool{
					"executorA": true,
					"executorB": true,
				},
				ReplicationFactor: 1,
				TallyGasLimit:     150000000000000,
				ExecGasLimit:      300000000000000,
				PostedGasPrice:    math.NewInt(1e5),
			},
			tallyGasLimit:   types.DefaultMaxTallyGasLimit,
			baseGasCost:     types.DefaultBaseGasCost,
			fallbackGasCost: 5e12,
			expectedExecutorRewards: []*types.DistributionExecutorReward{
				{
					Identity: "executorA",
					Amount:   math.NewIntFromUint64(5e17),
				},
				{
					Identity: "executorB",
					Amount:   math.NewIntFromUint64(5e17),
				},
			},
		},
		{
			name: "if a reveal is present, only reward the executor who revealed",
			dataRequest: &types.DataRequest{
				Commits: map[string][]byte{
					"executorA": []byte("commit"),
					"executorB": []byte("commit"),
					"executorC": []byte("commit"),
				},
				Reveals: map[string]bool{
					"executorB": true,
					"executorC": true,
				},
				ReplicationFactor: 1,
				TallyGasLimit:     150000000000000,
				ExecGasLimit:      300000000000000,
				PostedGasPrice:    math.NewInt(1e5),
			},
			tallyGasLimit:   types.DefaultMaxTallyGasLimit,
			baseGasCost:     types.DefaultBaseGasCost,
			fallbackGasCost: 1,
			expectedExecutorRewards: []*types.DistributionExecutorReward{
				{
					Identity: "executorB",
					Amount:   math.NewIntFromUint64(1e5),
				},
				{
					Identity: "executorC",
					Amount:   math.NewIntFromUint64(1e5),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gasMeter := types.NewGasMeter(tc.dataRequest, tc.tallyGasLimit, tc.baseGasCost)
			keeper.MeterExecutorGasFallback(*tc.dataRequest, tc.fallbackGasCost, gasMeter)
			dists := gasMeter.ReadGasMeter(ctx, "drID", 1, types.DefaultBurnRatio)

			require.Len(t, dists, len(tc.expectedExecutorRewards)+1)

			require.Nil(t, dists[0].ExecutorReward)
			require.Nil(t, dists[0].DataProxyReward)
			require.Equal(t, dists[0].Burn.Amount.String(), fmt.Sprintf("%d", tc.baseGasCost*tc.dataRequest.PostedGasPrice.Uint64()))

			for i, expected := range tc.expectedExecutorRewards {
				require.Nil(t, dists[i+1].Burn)
				require.Nil(t, dists[i+1].DataProxyReward)
				require.Equal(t, expected, dists[i+1].ExecutorReward)
			}
		})
	}
}

// Encountered this scenario on a testnet.
func TestReducedPayoutWithProxies(t *testing.T) {
	fixture := testutil.InitFixture(t)

	// Set up data proxies.
	proxyPubKey1, proxyPubKey2 := "03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0", "020173bd90e73c5f8576b3141c53aa9959b10a1daf1bc9c0ccf0a942932c703dec"
	proxyPayoutAddr1, proxyPayoutAddr2 := "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", "seda149sewl80wccuzhhukxgn2jg4kcun02d8qclwkt"
	err := fixture.SetDataProxyConfig(proxyPubKey1, proxyPayoutAddr1, sdk.NewCoin(testutil.BondDenom, math.NewInt(1000000000000000000)))
	require.NoError(t, err)
	err = fixture.SetDataProxyConfig(proxyPubKey2, proxyPayoutAddr2, sdk.NewCoin(testutil.BondDenom, math.NewInt(10000000000000)))
	require.NoError(t, err)

	// Scenario: 4 data proxy calls (3 to the same proxy, 1 to a different proxy), replication factor = 1.
	gasMeter := types.NewGasMeter(
		&types.DataRequest{
			TallyGasLimit:  150000000000000,
			ExecGasLimit:   300000000000000,
			PostedGasPrice: math.NewInt(100000),
		},
		types.DefaultMaxTallyGasLimit,
		types.DefaultBaseGasCost,
	)

	fixture.CoreKeeper.MeterProxyGas(
		fixture.Context(),
		[]string{proxyPubKey2, proxyPubKey1, proxyPubKey1, proxyPubKey1},
		1, gasMeter,
	)

	tallyRes := keeper.TallyResult{
		Reveals: []types.Reveal{
			{Executor: "020c4fe9e5063e7b5051284423089682082cf085a3b8f9e86bdb30407d761efc49"},
		},
		GasMeter:   gasMeter,
		GasReports: []uint64{81644889168750},
		FilterResult: keeper.FilterResult{
			ProxyPubKeys: []string{proxyPubKey1, proxyPubKey2},
			Outliers:     []bool{false},
		},
		ReplicationFactor: 1,
	}
	tallyRes.MeterExecutorGasUniform()

	require.Equalf(t, uint64(81644889168750), gasMeter.ExecutionGasUsed(), "expected exec gas used %d, got %d", 81644889168750, gasMeter.ExecutionGasUsed())
	require.Equalf(t, uint64(1000000000000), gasMeter.TallyGasUsed(), "expected tally gas used %d, got %d", 1000000100000, gasMeter.TallyGasUsed())

	dists := gasMeter.ReadGasMeter(fixture.Context(), "1", 1, types.DefaultBurnRatio)

	require.Len(t, dists, 6)

	require.Equal(t, "100000000000000000", dists[0].Burn.Amount.String(), "Burn amount is incorrect")
	require.Equal(t, "1000000000000000000", dists[1].DataProxyReward.Amount.String(), "Data proxy 1 (...2b0) did not receive correct payout")
	require.Equal(t, proxyPubKey1, dists[1].DataProxyReward.PublicKey, "Data proxy 1 (...2b0) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr1, dists[1].DataProxyReward.PayoutAddress, "Data proxy 1 (...2b0) did not receive correct payout address")

	require.Equal(t, "1000000000000000000", dists[2].DataProxyReward.Amount.String(), "Data proxy 1 (...2b0) did not receive correct payout")
	require.Equal(t, proxyPubKey1, dists[2].DataProxyReward.PublicKey, "Data proxy 1 (...2b0) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr1, dists[2].DataProxyReward.PayoutAddress, "Data proxy 1 (...2b0) did not receive correct payout address")

	require.Equal(t, "1000000000000000000", dists[3].DataProxyReward.Amount.String(), "Data proxy 1 (...2b0) did not receive correct payout")
	require.Equal(t, proxyPubKey1, dists[3].DataProxyReward.PublicKey, "Data proxy 1 (...2b0) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr1, dists[3].DataProxyReward.PayoutAddress, "Data proxy 1 (...2b0) did not receive correct payout address")

	require.Equal(t, "10000000000000", dists[4].DataProxyReward.Amount.String(), "Data proxy 2 (...dec) did not receive correct payout")
	require.Equal(t, proxyPubKey2, dists[4].DataProxyReward.PublicKey, "Data proxy 2 (...dec) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr2, dists[4].DataProxyReward.PayoutAddress, "Data proxy 2 (...dec) did not receive correct payout address")

	require.Equal(t, "5164478916875000000", dists[5].ExecutorReward.Amount.String(), "Executor did not receive correct payout")

	// Test with reduced payout mode.
	gasMeter.SetReducedPayoutMode()

	// Sanity check that the gas meter is not affected by the reduced payout mode.
	require.Equalf(t, uint64(81644889168750), gasMeter.ExecutionGasUsed(), "expected exec gas used %d, got %d", 81644889168750, gasMeter.ExecutionGasUsed())
	require.Equalf(t, uint64(1000000000000), gasMeter.TallyGasUsed(), "expected tally gas used %d, got %d", 1000000100000, gasMeter.TallyGasUsed())

	distsReduced := gasMeter.ReadGasMeter(fixture.Context(), "1", 1, types.DefaultBurnRatio)

	require.Equal(t, "1132895783375000000", distsReduced[0].Burn.Amount.String(), "Burn amount is incorrect")

	require.Equal(t, "1000000000000000000", distsReduced[1].DataProxyReward.Amount.String(), "Data proxy 1 (...2b0) did not receive correct payout")
	require.Equal(t, proxyPubKey1, distsReduced[1].DataProxyReward.PublicKey, "Data proxy 1 (...2b0) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr1, distsReduced[1].DataProxyReward.PayoutAddress, "Data proxy 1 (...2b0) did not receive correct payout address")

	require.Equal(t, "1000000000000000000", distsReduced[2].DataProxyReward.Amount.String(), "Data proxy 1 (...2b0) did not receive correct payout")
	require.Equal(t, proxyPubKey1, distsReduced[2].DataProxyReward.PublicKey, "Data proxy 1 (...2b0) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr1, distsReduced[2].DataProxyReward.PayoutAddress, "Data proxy 1 (...2b0) did not receive correct payout address")

	require.Equal(t, "1000000000000000000", distsReduced[3].DataProxyReward.Amount.String(), "Data proxy 1 (...2b0) did not receive correct payout")
	require.Equal(t, proxyPubKey1, distsReduced[3].DataProxyReward.PublicKey, "Data proxy 1 (...2b0) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr1, distsReduced[3].DataProxyReward.PayoutAddress, "Data proxy 1 (...2b0) did not receive correct payout address")

	require.Equal(t, "10000000000000", distsReduced[4].DataProxyReward.Amount.String(), "Data proxy 2 (...dec) did not receive correct payout")
	require.Equal(t, proxyPubKey2, distsReduced[4].DataProxyReward.PublicKey, "Data proxy 2 (...dec) did not receive correct public key")
	require.Equal(t, proxyPayoutAddr2, distsReduced[4].DataProxyReward.PayoutAddress, "Data proxy 2 (...dec) did not receive correct payout address")

	require.Equal(t, "4131583133500000000", distsReduced[5].ExecutorReward.Amount.String(), "Executor did not receive correct reduced payout")

	// Sanity check that the difference between the two distributions is the same as the reduced payout.
	require.Equal(t, distsReduced[0].Burn.Amount.Sub(dists[0].Burn.Amount).String(), dists[5].ExecutorReward.Amount.Sub(distsReduced[5].ExecutorReward.Amount).String(), "Difference between burn and executor reward is not the same as the reduced payout")
}
