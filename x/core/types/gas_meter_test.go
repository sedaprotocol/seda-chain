package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func TestGasMeter_TallyGas(t *testing.T) {
	ctx := sdktestutil.DefaultContext(
		storetypes.NewKVStoreKey(types.StoreKey),
		storetypes.NewTransientStoreKey("transient_test"),
	)
	gasPrice := uint64(1000)

	tests := []struct {
		name              string
		amount            uint64
		tallyGasLimit     uint64 // requestor-specified
		maxTallyGasLimit  uint64 // parameter
		baseGasCost       uint64 // parameter
		expectedUsed      uint64
		expectedRemaining uint64
		expectedExhausted bool // true if the tally gas limit is not sufficient to cover the amount
	}{
		{
			name:              "normal consumption",
			amount:            5000,
			tallyGasLimit:     10000,
			maxTallyGasLimit:  20000,
			baseGasCost:       500,
			expectedUsed:      5500,
			expectedRemaining: 4500,
			expectedExhausted: false,
		},
		{
			name:              "exceeds remaining",
			amount:            10000,
			tallyGasLimit:     10000,
			maxTallyGasLimit:  20000,
			baseGasCost:       1000,
			expectedUsed:      10000,
			expectedRemaining: 0,
			expectedExhausted: true,
		},
		{
			name:              "exact consumption",
			amount:            7800,
			tallyGasLimit:     9000,
			maxTallyGasLimit:  10000,
			baseGasCost:       1200,
			expectedUsed:      9000,
			expectedRemaining: 0,
			expectedExhausted: false,
		},
		{
			name:              "base gas cost at max tally gas limit",
			amount:            5000,
			tallyGasLimit:     9000,
			maxTallyGasLimit:  10000,
			baseGasCost:       10000,
			expectedUsed:      9000,
			expectedRemaining: 0,
			expectedExhausted: true,
		},
		{
			name:              "base gas cost exceeds max tally gas limit",
			amount:            5000,
			tallyGasLimit:     9000,
			maxTallyGasLimit:  10000,
			baseGasCost:       15000,
			expectedUsed:      9000,
			expectedRemaining: 0,
			expectedExhausted: true,
		},
		{
			name:              "tally gas limit exceeds max tally gas limit",
			amount:            5000,
			tallyGasLimit:     9000,
			maxTallyGasLimit:  7500,
			baseGasCost:       50,
			expectedUsed:      5050,
			expectedRemaining: 2450,
			expectedExhausted: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gasMeter := NewGasMeter(&DataRequest{
				TallyGasLimit:  tc.tallyGasLimit,
				ExecGasLimit:   10000,
				PostedGasPrice: math.NewIntFromUint64(gasPrice),
				Poster:         "seda1test",
				Escrow:         math.NewIntFromUint64(1000000),
			}, tc.maxTallyGasLimit, tc.baseGasCost)

			exhausted := gasMeter.ConsumeTallyGas(tc.amount)
			require.Equal(t, tc.expectedRemaining, gasMeter.tallyGasRemaining)
			require.Equal(t, tc.expectedExhausted, exhausted)
			require.Equal(t, tc.expectedUsed, gasMeter.TallyGasUsed())
			if tc.expectedExhausted {
				require.Equal(t, min(tc.tallyGasLimit, tc.maxTallyGasLimit), gasMeter.TallyGasUsed())
			}

			dists := gasMeter.ReadGasMeter(ctx, "drID", 123, DefaultBurnRatio)
			require.Equal(t, 1, len(dists))
			require.Equal(t, math.NewIntFromUint64(tc.expectedUsed*gasPrice), dists[0].Burn.Amount)
			require.Nil(t, dists[0].ExecutorReward)
			require.Nil(t, dists[0].DataProxyReward)
		})
	}
}

func TestGasMeter_ExecutionGas(t *testing.T) {
	ctx := sdktestutil.DefaultContext(
		storetypes.NewKVStoreKey(types.StoreKey),
		storetypes.NewTransientStoreKey("transient_test"),
	)

	execGasLimit := uint64(10000)
	gasPrice := uint64(500)

	gasMeter := NewGasMeter(&DataRequest{
		TallyGasLimit:  DefaultMaxTallyGasLimit,
		ExecGasLimit:   execGasLimit,
		PostedGasPrice: math.NewIntFromUint64(gasPrice),
		Poster:         "seda1test",
		Escrow:         math.NewIntFromUint64(1000000),
	}, DefaultMaxTallyGasLimit, DefaultBaseGasCost)

	var expectedDists = []Distribution{
		{
			Burn: &DistributionBurn{
				Amount: math.NewIntFromUint64(DefaultBaseGasCost * gasPrice),
			},
		},
	}

	proxyCount := 0
	tests := []struct {
		name     string
		forProxy *struct {
			proxyPubKey       string
			payoutAddr        string
			replicationFactor uint64
			amount            uint64
		}
		forExecutor *struct {
			executorPubKey string
			amount         uint64
		}
		expectedConsumed  uint64
		expectedRemaining uint64
	}{
		{
			name: "consume exec gas for executor1",
			forExecutor: &struct {
				executorPubKey string
				amount         uint64
			}{
				executorPubKey: "executor1",
				amount:         3000,
			},
			expectedConsumed:  3000,
			expectedRemaining: 7000,
		},
		{
			name: "consume exec gas for proxy1",
			forProxy: &struct {
				proxyPubKey       string
				payoutAddr        string
				replicationFactor uint64
				amount            uint64
			}{
				proxyPubKey:       "proxy1",
				payoutAddr:        "addr1",
				replicationFactor: 3,
				amount:            500,
			},
			expectedConsumed:  1500,
			expectedRemaining: 5500, // 7000 - amount * RF
		},
		{
			name: "consume exec gas for proxy2",
			forProxy: &struct {
				proxyPubKey       string
				payoutAddr        string
				replicationFactor uint64
				amount            uint64
			}{
				proxyPubKey:       "proxy2",
				payoutAddr:        "addr2",
				replicationFactor: 3,
				amount:            300,
			},
			expectedRemaining: 4600, // 5500 - amount * RF
		},
		{
			name: "exhausts remaining exec gas for executor2",
			forExecutor: &struct {
				executorPubKey string
				amount         uint64
			}{
				executorPubKey: "executor2",
				amount:         8000,
			},
			expectedConsumed:  4600,
			expectedRemaining: 0, // exhausted
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if (tc.forExecutor != nil && tc.forProxy != nil) || (tc.forExecutor == nil && tc.forProxy == nil) {
				t.Fatal("specify either forExecutor or forProxy per iteration")
			}

			if tc.forExecutor != nil {
				gasMeter.ConsumeExecGasForExecutor(tc.forExecutor.executorPubKey, tc.forExecutor.amount)
			} else {
				gasMeter.ConsumeExecGasForProxy(tc.forProxy.proxyPubKey, tc.forProxy.payoutAddr, tc.forProxy.replicationFactor, tc.forProxy.amount)
				proxyCount++
			}

			require.Equal(t, tc.expectedRemaining, gasMeter.execGasRemaining)

			dists := gasMeter.ReadGasMeter(ctx, "drID", 123, DefaultBurnRatio)

			require.Equal(t, i+2, len(dists))

			var expectedDist Distribution
			if tc.forExecutor != nil {
				expectedDist.ExecutorReward = &DistributionExecutorReward{
					Identity: tc.forExecutor.executorPubKey,
					Amount:   math.NewIntFromUint64(tc.forExecutor.amount * gasPrice),
				}
			} else {
				expectedDist.DataProxyReward = &DistributionDataProxyReward{
					PublicKey:     tc.forProxy.proxyPubKey,
					PayoutAddress: tc.forProxy.payoutAddr,
					Amount:        math.NewIntFromUint64(tc.forProxy.amount * tc.forProxy.replicationFactor * gasPrice),
				}
			}
			expectedDists = append(expectedDists, expectedDist)

			require.ElementsMatch(t, dists, expectedDists, "expected dists %v, got %v", expectedDists, dists)
		})
	}
}
