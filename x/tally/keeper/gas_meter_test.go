package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func FuzzGasMetering(f *testing.F) {
	fixture := initFixture(f)

	// Prepare fixed parameters of the fuzz test.
	defaultParams := types.DefaultParams()
	err := fixture.tallyKeeper.SetParams(fixture.Context(), defaultParams)
	require.NoError(f, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm2(), fixture.Context().BlockTime())
	err = fixture.wasmStorageKeeper.OracleProgram.Set(fixture.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(f, err)

	filterInput, err := hex.DecodeString("01000000000000000D242E726573756C742E74657874") // mode, json_path = $.result.text
	require.NoError(f, err)

	proxyPubKeys := []string{"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8", "2a4c8d5b3ef9a1c7d6b430e78f9dcc2a2a1440f9"}
	pubKeyToPayoutAddr := map[string]string{
		proxyPubKeys[0]: "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f",
		proxyPubKeys[1]: "seda149sewl80wccuzhhukxgn2jg4kcun02d8qclwkt",
	}
	proxyFee := sdk.NewCoin(bondDenom, math.NewIntWithDecimal(1, 21))
	expProxyGasUsed := map[string]math.Int{
		pubKeyToPayoutAddr[proxyPubKeys[0]]: math.NewInt(10000), // = RF * proxyFee / gasPrice
		pubKeyToPayoutAddr[proxyPubKeys[1]]: math.NewInt(10000), // = RF * proxyFee / gasPrice
	}
	for _, pk := range proxyPubKeys {
		err = fixture.SetDataProxyConfig(pk, pubKeyToPayoutAddr[pk], proxyFee)
		require.NoError(f, err)
	}

	execGasLimit := uint64(1e11) // ^uint64(0)
	tallyGasLimit := uint64(types.DefaultMaxTallyGasLimit)

	f.Fuzz(func(t *testing.T, g0, g1, g2, g3, g4, g5, g6, g7, g8, g9 uint64) {
		t.Log(g0, g1, g2, g3, g4, g5, g6, g7, g8, g9)

		reveals := map[string]types.RevealBody{
			"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g0, ProxyPubKeys: proxyPubKeys},
			"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g1, ProxyPubKeys: proxyPubKeys},
			"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g2, ProxyPubKeys: proxyPubKeys},
			"d": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g3, ProxyPubKeys: proxyPubKeys},
			"e": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g4, ProxyPubKeys: proxyPubKeys},
			"f": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g5, ProxyPubKeys: proxyPubKeys},
			"g": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g6, ProxyPubKeys: proxyPubKeys},
			"h": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g7, ProxyPubKeys: proxyPubKeys},
			"i": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g8, ProxyPubKeys: proxyPubKeys},
			"j": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: g9, ProxyPubKeys: proxyPubKeys},
		}

		// To avoid commit timeout (no other effect intended)
		commits := make(map[string][]byte)
		for i := range len(reveals) {
			commits[fmt.Sprintf("executor-%d", i)] = []byte{}
		}

		tallyRes, dataRes, processedReqs, err := fixture.tallyKeeper.ProcessTallies(
			fixture.Context(),
			[]types.Request{{
				Commits:           commits,
				Reveals:           reveals,
				ReplicationFactor: uint16(len(reveals)),
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
				PostedGasPrice:    "1000000000000000000",
				ExecGasLimit:      execGasLimit,
				TallyGasLimit:     tallyGasLimit,
				TallyProgramID:    hex.EncodeToString(tallyProgram.Hash),
			}},
			types.DefaultParams(), false,
		)
		require.NoError(t, err)

		require.Equal(t, 1, len(tallyRes))
		require.Equal(t, 1, len(dataRes))
		require.Equal(t, 1, len(processedReqs))

		gasMeter := tallyRes[0].GasMeter

		// Check executor gas used.
		sumExec := math.NewInt(0)
		for _, exec := range gasMeter.GetExecutorGasUsed() {
			sumExec = sumExec.Add(exec.Amount)
		}

		// Check proxy gas used.
		for _, proxy := range gasMeter.GetProxyGasUsed("dummy-request-id", fixture.Context().BlockHeight()) {
			require.Equal(t,
				expProxyGasUsed[proxy.PayoutAddress].String(),
				proxy.Amount.String(),
			)
			sumExec = sumExec.Add(proxy.Amount)
		}

		sumExec = sumExec.Add(math.NewIntFromUint64(gasMeter.RemainingExecGas()))
		require.Equal(t, sumExec.String(), strconv.FormatUint(execGasLimit, 10))

		tallySum := math.NewIntFromUint64(gasMeter.TallyGasUsed())
		tallySum = tallySum.Add(math.NewIntFromUint64(gasMeter.RemainingTallyGas()))
		require.Equal(t, tallySum.String(), strconv.FormatUint(tallyGasLimit, 10))

		dists := fixture.tallyKeeper.DistributionsFromGasMeter(fixture.Context(), "1", 1, gasMeter, types.DefaultBurnRatio)
		require.Len(t, dists, 13)

		totalDist := math.NewInt(0)
		totalExecDist := math.NewInt(0)
		burn := math.NewInt(0)
		for _, dist := range dists {
			if dist.Burn != nil {
				burn = burn.Add(dist.Burn.Amount)
				totalDist = totalDist.Add(dist.Burn.Amount)
			}
			if dist.DataProxyReward != nil {
				totalDist = totalDist.Add(dist.DataProxyReward.Amount)
			}
			if dist.ExecutorReward != nil {
				totalExecDist = totalExecDist.Add(dist.ExecutorReward.Amount)
				totalDist = totalDist.Add(dist.ExecutorReward.Amount)
			}
		}

		totalGasPayed := totalDist.Quo(gasMeter.GasPrice())

		require.True(t, totalGasPayed.LTE(sumExec.Add((tallySum))), "total gas paid is not less than or equal to the sum of exec and tally gas used")

		gasMeter.SetReducedPayoutMode()
		distsReduced := fixture.tallyKeeper.DistributionsFromGasMeter(fixture.Context(), "1", 1, gasMeter, types.DefaultBurnRatio)
		totalDistReduced := math.NewInt(0)
		burnReduced := math.NewInt(0)
		for _, dist := range distsReduced {
			if dist.Burn != nil {
				burnReduced = burnReduced.Add(dist.Burn.Amount)
				totalDistReduced = totalDistReduced.Add(dist.Burn.Amount)
			}
			if dist.DataProxyReward != nil {
				totalDistReduced = totalDistReduced.Add(dist.DataProxyReward.Amount)
			}
			if dist.ExecutorReward != nil {
				totalDistReduced = totalDistReduced.Add(dist.ExecutorReward.Amount)
			}
		}

		totalGasPayedReduced := totalDistReduced.Quo(gasMeter.GasPrice())

		require.Equal(t, totalGasPayedReduced.String(), totalGasPayed.String(), "total gas paid in reduced mode is not equal to the total gas paid in normal mode")
		if totalExecDist.Equal(math.NewInt(0)) {
			require.True(t, burnReduced.Equal(burn), "burn amount in reduced mode is not equal to the burn amount in normal mode when there is no exec gas used")
		} else {
			require.True(t, burnReduced.GT(burn), "burn amount in reduced mode is equal to the burn amount in normal mode when there is exec gas used")
		}
	})
}

// Encountered this scenario on a testnet.
func TestReducedPayoutWithProxies(t *testing.T) {
	fixture := initFixture(t)

	// Set up data proxies.
	proxyPubKey1, proxyPubKey2 := "03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0", "020173bd90e73c5f8576b3141c53aa9959b10a1daf1bc9c0ccf0a942932c703dec"
	proxyPayoutAddr1, proxyPayoutAddr2 := "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", "seda149sewl80wccuzhhukxgn2jg4kcun02d8qclwkt"
	err := fixture.SetDataProxyConfig(proxyPubKey1, proxyPayoutAddr1, sdk.NewCoin(bondDenom, math.NewInt(1000000000000000000)))
	require.NoError(t, err)
	err = fixture.SetDataProxyConfig(proxyPubKey2, proxyPayoutAddr2, sdk.NewCoin(bondDenom, math.NewInt(10000000000000)))
	require.NoError(t, err)

	// Scenario: 4 data proxy calls (3 to the same proxy, 1 to a different proxy), replication factor = 1.
	gasMeter := types.NewGasMeter(150000000000000, 300000000000000, types.DefaultMaxTallyGasLimit, math.NewInt(100000), types.DefaultGasCostBase)
	fixture.tallyKeeper.MeterProxyGas(fixture.Context(), []string{"020173bd90e73c5f8576b3141c53aa9959b10a1daf1bc9c0ccf0a942932c703dec", "03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0", "03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0", "03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0"}, 1, gasMeter)

	tallyRes := types.TallyResult{
		Reveals: []types.Reveal{
			{Executor: "020c4fe9e5063e7b5051284423089682082cf085a3b8f9e86bdb30407d761efc49"},
		},
		GasMeter:   gasMeter,
		GasReports: []uint64{81644889168750},
		FilterResult: types.FilterResult{
			ProxyPubKeys: []string{proxyPubKey1, proxyPubKey2},
			Outliers:     []bool{false},
		},
		ReplicationFactor: 1,
	}
	tallyRes.MeterExecutorGasUniform()

	require.Equalf(t, uint64(81644889168750), gasMeter.ExecutionGasUsed(), "expected exec gas used %d, got %d", 81644889168750, gasMeter.ExecutionGasUsed())
	require.Equalf(t, uint64(1000000000000), gasMeter.TallyGasUsed(), "expected tally gas used %d, got %d", 1000000100000, gasMeter.TallyGasUsed())

	dists := fixture.tallyKeeper.DistributionsFromGasMeter(fixture.Context(), "1", 1, gasMeter, types.DefaultBurnRatio)

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

	distsReduced := fixture.tallyKeeper.DistributionsFromGasMeter(fixture.Context(), "1", 1, gasMeter, types.DefaultBurnRatio)

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
