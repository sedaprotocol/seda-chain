package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	"github.com/sedaprotocol/seda-chain/x/tally/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func FuzzGasMetering(f *testing.F) {
	fixture := initFixture(f)

	// Prepare fixed parameters of the fuzz test.
	defaultParams := types.DefaultParams()
	err := fixture.tallyKeeper.SetParams(fixture.Context(), defaultParams)
	require.NoError(f, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm2(), fixture.Context().BlockTime())
	err = fixture.wasmStorageKeeper.OracleProgram.Set(fixture.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(f, err)

	gasPriceStr := "1000000000000000000" // 1e18
	gasPrice, ok := math.NewIntFromString(gasPriceStr)
	require.True(f, ok)

	filterInput, err := hex.DecodeString("01000000000000000D242E726573756C742E74657874") // mode, json_path = $.result.text
	require.NoError(f, err)

	proxyPubKeys := []string{"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8", "2a4c8d5b3ef9a1c7d6b430e78f9dcc2a2a1440f9"}
	proxyFee := sdk.NewCoin(bondDenom, math.NewIntWithDecimal(1, 21))
	expProxyGasUsed := map[string]math.Int{
		"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8": math.NewInt(10000), // = RF * proxyFee / gasPrice
		"2a4c8d5b3ef9a1c7d6b430e78f9dcc2a2a1440f9": math.NewInt(10000), // = RF * proxyFee / gasPrice
	}
	for _, pk := range proxyPubKeys {
		pkBytes, err := hex.DecodeString(pk)
		if err == nil {
			err := fixture.dataProxyKeeper.SetDataProxyConfig(fixture.Context(), pkBytes,
				dataproxytypes.ProxyConfig{
					Fee: &proxyFee,
				},
			)
			require.NoError(f, err)
		}
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

		gasMeter := types.NewGasMeter(tallyGasLimit, execGasLimit, types.DefaultMaxTallyGasLimit, gasPrice, types.DefaultGasCostBase)

		fixture.tallyKeeper.FilterAndTally(
			fixture.Context(),
			types.Request{
				Reveals:           reveals,
				ReplicationFactor: uint16(len(reveals)),
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
				GasPrice:          gasPriceStr,
				ExecGasLimit:      execGasLimit,
				TallyGasLimit:     tallyGasLimit,
				TallyProgramID:    hex.EncodeToString(tallyProgram.Hash),
			}, types.DefaultParams(), gasMeter)

		// Check executor gas used.
		sumExec := math.NewInt(0)
		for _, exec := range gasMeter.Executors {
			sumExec = sumExec.Add(exec.Amount)
		}

		// Check proxy gas used.
		for _, proxy := range gasMeter.Proxies {
			require.Equal(t,
				expProxyGasUsed[hex.EncodeToString(proxy.PublicKey)].String(),
				proxy.Amount.String(),
			)
			sumExec = sumExec.Add(proxy.Amount)
		}

		sumExec = sumExec.Add(math.NewIntFromUint64(gasMeter.RemainingExecGas()))
		require.Equal(t, sumExec.String(), strconv.FormatUint(execGasLimit, 10))

		tallySum := math.NewIntFromUint64(gasMeter.TallyGasUsed())
		tallySum = tallySum.Add(math.NewIntFromUint64(gasMeter.RemainingTallyGas()))
		require.Equal(t, tallySum.String(), strconv.FormatUint(tallyGasLimit, 10))
	})
}
