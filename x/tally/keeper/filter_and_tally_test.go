package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	"github.com/sedaprotocol/seda-chain/x/tally/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func TestFilterAndTally(t *testing.T) {
	f := initFixture(t)

	defaultParams := types.DefaultParams()
	err := f.tallyKeeper.SetParams(f.Context(), defaultParams)
	require.NoError(t, err)

	tests := []struct {
		name              string
		tallyInputAsHex   string
		outliers          []bool
		reveals           map[string]types.RevealBody
		replicationFactor uint16
		consensus         bool
		consPubKeys       []string // expected proxy public keys in basic consensus
		tallyGasUsed      uint64
		exitCode          uint32
		filterErr         error
	}{
		{
			name:            "None filter - One reveal missing",
			tallyInputAsHex: "00",
			outliers:        []bool{false, false, false, false},
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				"d": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
			},
			replicationFactor: 5,
			consensus:         true,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase + defaultParams.FilterGasCostNone,
			exitCode:          types.TallyExitCodeExecError, // since tally program does not exist
			filterErr:         nil,
		},
		{
			name:            "None filter - Four reveals missing",
			tallyInputAsHex: "00",
			outliers:        []bool{false},
			reveals: map[string]types.RevealBody{
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
			},
			replicationFactor: 5,
			consensus:         false,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase,
			exitCode:          types.TallyExitCodeFilterError,
			filterErr:         types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - One reveal missing",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, false, false, false},
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				"d": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
			},
			replicationFactor: 5,
			consensus:         true,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*5,
			exitCode:          types.TallyExitCodeExecError, // since tally program does not exist
			filterErr:         nil,
		},
		{
			name:            "Mode filter - Four reveals missing",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false},
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
			},
			replicationFactor: 5,
			consensus:         false,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase,
			exitCode:          types.TallyExitCodeFilterError,
			filterErr:         types.ErrNoBasicConsensus,
		},
		{
			name:            "Standard deviation filter - One reveal missing",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        []bool{false, false, false, false},
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": 5}}`},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": 6}}`},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": 4}}`},
				"d": {ExitCode: 0, Reveal: `{"result": {"text": 6}}`},
			},
			replicationFactor: 5,
			consensus:         true,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierStdDev*5,
			exitCode:          types.TallyExitCodeExecError, // since tally program does not exist
			filterErr:         nil,
		},
		{
			name:            "Standard deviation filter - Four reveals missing",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        []bool{false},
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": 5}}`},
			},
			replicationFactor: 5,
			consensus:         false,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase,
			exitCode:          types.TallyExitCodeFilterError,
			filterErr:         types.ErrNoBasicConsensus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterInput, err := hex.DecodeString(tt.tallyInputAsHex)
			require.NoError(t, err)

			reveals := make(map[string]types.RevealBody)
			for k, v := range tt.reveals {
				revealBody := v
				revealBody.Reveal = base64.StdEncoding.EncodeToString([]byte(v.Reveal))
				revealBody.GasUsed = v.GasUsed
				reveals[k] = revealBody
			}

			gasMeter := types.NewGasMeter(1e13, 0, types.DefaultMaxTallyGasLimit, math.NewIntWithDecimal(1, 18), types.DefaultGasCostBase)
			filterRes, tallyRes := f.tallyKeeper.FilterAndTally(f.Context(), types.Request{
				Reveals:           reveals,
				ReplicationFactor: tt.replicationFactor,
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
				GasPrice:          "1000000000000000000", // 1e18
				ExecGasLimit:      100000,
			}, types.DefaultParams(), gasMeter)
			require.NoError(t, err)

			require.Equal(t, tt.outliers, filterRes.Outliers)
			require.Equal(t, tt.tallyGasUsed, gasMeter.TallyGasUsed())
			require.Equal(t, tt.consensus, filterRes.Consensus)
			require.Equal(t, tt.consensus, tallyRes.Consensus)
			require.Equal(t, tt.exitCode, tallyRes.ExitCode)
			if tt.filterErr != nil {
				require.Equal(t, []byte(tt.filterErr.Error()), tallyRes.Result)
			}

			if tt.consPubKeys == nil {
				require.Nil(t, nil, tallyRes.ProxyPubKeys)
			} else {
				for _, pk := range tt.consPubKeys {
					require.Contains(t, tallyRes.ProxyPubKeys, pk)
				}
			}
		})
	}
}

func TestExecutorPayout(t *testing.T) {
	f := initFixture(t)

	defaultParams := types.DefaultParams()
	err := f.tallyKeeper.SetParams(f.Context(), defaultParams)
	require.NoError(t, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm2(), f.Context().BlockTime())
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	pubKeys := []string{"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8", "2a4c8d5b3ef9a1c7d6b430e78f9dcc2a2a1440f9"}
	pubKeyToPayoutAddr := map[string]string{
		pubKeys[0]: "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f",
		pubKeys[1]: "seda149sewl80wccuzhhukxgn2jg4kcun02d8qclwkt",
	}

	tests := []struct {
		name              string
		tallyInputAsHex   string
		reveals           map[string]types.RevealBody
		replicationFactor uint16
		execGasLimit      uint64
		expExecGasUsed    uint64
		expReducedPayout  bool
		expExecutorGas    map[string]math.Int
		expProxyGas       map[string]math.Int
	}{
		{
			name:            "Uniform gas reporting",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    90000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(30000),
				"b": math.NewInt(30000),
				"c": math.NewInt(30000),
			},
		},
		{
			name:            "Uniform gas reporting beyond execGasLimit",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
			},
			replicationFactor: 3,
			execGasLimit:      60000,
			expExecGasUsed:    60000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(20000),
				"b": math.NewInt(20000),
				"c": math.NewInt(20000),
			},
		},
		{
			name:            "Uniform gas reporting (consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 30000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    90000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(30000),
				"b": math.NewInt(30000),
				"c": math.NewInt(30000),
			},
		},
		{
			name:            "Uniform gas reporting (mode consensus in error)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 1, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"b": {ExitCode: 1, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 30000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    90000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(30000),
				"b": math.NewInt(30000),
				"c": math.NewInt(30000),
			},
		},
		{
			name:            "Uniform gas reporting (mode no consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 20000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
			},
			replicationFactor: 3,
			execGasLimit:      60000,
			expExecGasUsed:    60000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(20000),
				"b": math.NewInt(20000),
				"c": math.NewInt(20000),
			},
			expReducedPayout: true,
		},
		{
			name:            "Uniform gas reporting with low gas limit (mode no consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 20000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
			},
			replicationFactor: 3,
			execGasLimit:      1000,
			expExecGasUsed:    999,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(333),
				"b": math.NewInt(333),
				"c": math.NewInt(333),
			},
			expReducedPayout: true,
		},
		{
			name:            "Divergent gas reporting with shares (lowest_report*2 > median_report)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    43448 + 23275 + 23275,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(43448),
				"b": math.NewInt(23275),
				"c": math.NewInt(23275),
			},
		},
		{
			name:            "Divergent gas reporting without shares (lowest_report*2 < median_report)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 20000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    56000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(16000),
				"b": math.NewInt(20000),
				"c": math.NewInt(20000),
			},
		},
		{
			name:            "Divergent gas reporting without shares (lowest_report*2 == median_report)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 10000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 20000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    60000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(20000),
				"b": math.NewInt(20000),
				"c": math.NewInt(20000),
			},
		},
		{
			name:            "Divergent gas reporting (mode no consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    56000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(16000),
				"b": math.NewInt(20000),
				"c": math.NewInt(20000),
			},
			expReducedPayout: true,
		},
		{
			name:            "Divergent gas reporting with low gas limit and no shares (mode no consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 1},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      1000,
			expExecGasUsed:    668,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(2),
				"b": math.NewInt(333),
				"c": math.NewInt(333),
			},
			expReducedPayout: true,
		},
		{
			name:            "Divergent gas reporting with low gas limit and shares (mode no consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      1000,
			expExecGasUsed:    997,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(499),
				"b": math.NewInt(249),
				"c": math.NewInt(249),
			},
			expReducedPayout: true,
		},
		{
			name:            "Divergent gas reporting (mode no consensus, with 1 proxy)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{ // (7000, 19000, 34000) after subtracting proxy gas
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000, ProxyPubKeys: []string{pubKeys[0]}},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000, ProxyPubKeys: []string{pubKeys[0]}},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000, ProxyPubKeys: []string{pubKeys[0]}},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    55000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(14000),
				"b": math.NewInt(19000),
				"c": math.NewInt(19000),
			},
			expReducedPayout: true,
			expProxyGas: map[string]math.Int{
				pubKeyToPayoutAddr[pubKeys[0]]: math.NewInt(3000), // = RF * proxyFee / gasPrice
			},
		},
		{
			name:            "Divergent gas reporting (mode no consensus, with 2 proxies)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{ // (6000, 18000, 33000) after subtracting proxy gas
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000, ProxyPubKeys: []string{pubKeys[0], pubKeys[1]}},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000, ProxyPubKeys: []string{pubKeys[0], pubKeys[1]}},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000, ProxyPubKeys: []string{pubKeys[0]}},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    54000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(12000),
				"b": math.NewInt(18000),
				"c": math.NewInt(18000),
			},
			expReducedPayout: true,
			expProxyGas: map[string]math.Int{
				pubKeyToPayoutAddr[pubKeys[0]]: math.NewInt(3000), // = RF * proxyFee / gasPrice
				pubKeyToPayoutAddr[pubKeys[1]]: math.NewInt(3000), // = RF * proxyFee / gasPrice
			},
		},
		{
			name:            "Divergent gas reporting with low gas limit (mode no consensus, with 2 proxies)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{ // (0, 0, 0) after subtracting proxy gas and considering gas limit
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000, ProxyPubKeys: []string{pubKeys[0], pubKeys[1]}},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000, ProxyPubKeys: []string{pubKeys[0], pubKeys[1]}},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000, ProxyPubKeys: []string{pubKeys[0]}},
			},
			replicationFactor: 3,
			execGasLimit:      1000,
			expExecGasUsed:    999,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(0),
				"b": math.NewInt(0),
				"c": math.NewInt(0),
			},
			expReducedPayout: true,
			expProxyGas: map[string]math.Int{
				pubKeyToPayoutAddr[pubKeys[0]]: math.NewInt(999), // = RF * proxyFee / gasPrice (considering gas limit)
				pubKeyToPayoutAddr[pubKeys[1]]: math.NewInt(0),   // = RF * proxyFee / gasPrice (considering gas limit)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterInput, err := hex.DecodeString(tt.tallyInputAsHex)
			require.NoError(t, err)

			exp21, ok := math.NewIntFromString("1000000000000000000000") // 1e21
			require.True(t, ok)
			proxyFee := sdk.NewCoin(bondDenom, exp21)
			reveals := make(map[string]types.RevealBody)
			for k, v := range tt.reveals {
				revealBody := v
				revealBody.Reveal = base64.StdEncoding.EncodeToString([]byte(v.Reveal))
				revealBody.GasUsed = v.GasUsed
				reveals[k] = revealBody

				for _, pk := range v.ProxyPubKeys {
					pkBytes, err := hex.DecodeString(pk)
					if err == nil {
						err := f.dataProxyKeeper.SetDataProxyConfig(f.Context(), pkBytes,
							dataproxytypes.ProxyConfig{
								PayoutAddress: pubKeyToPayoutAddr[pk],
								Fee:           &proxyFee,
							},
						)
						require.NoError(t, err)
					}
				}
			}

			gasPriceStr := "1000000000000000000" // 1e18
			gasPrice, ok := math.NewIntFromString(gasPriceStr)
			require.True(t, ok)

			request := types.Request{
				Reveals:           reveals,
				ReplicationFactor: tt.replicationFactor,
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
				GasPrice:          gasPriceStr,
				TallyGasLimit:     types.DefaultMaxTallyGasLimit,
				ExecGasLimit:      tt.execGasLimit,
				TallyProgramID:    hex.EncodeToString(tallyProgram.Hash),
			}

			gasMeter := types.NewGasMeter(request.TallyGasLimit, request.ExecGasLimit, types.DefaultMaxTallyGasLimit, gasPrice, types.DefaultGasCostBase)
			_, tallyRes := f.tallyKeeper.FilterAndTally(f.Context(), request, types.DefaultParams(), gasMeter)
			require.NoError(t, err)

			for _, exec := range gasMeter.Executors {
				require.Equal(t,
					tt.expExecutorGas[exec.PublicKey].String(),
					exec.Amount.String(),
				)
			}
			require.Equal(t, tt.expReducedPayout, gasMeter.ReducedPayout)
			for _, proxy := range gasMeter.Proxies {
				require.Equal(t,
					tt.expProxyGas[proxy.PayoutAddress].String(),
					proxy.Amount.String(),
				)
			}
			require.Equal(t, tt.expExecGasUsed, tallyRes.ExecGasUsed)
			require.Equal(t, tt.expExecGasUsed, gasMeter.ExecutionGasUsed())
		})
	}
}
