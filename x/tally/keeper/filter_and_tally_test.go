package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"slices"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
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
		reveals           []types.RevealBody
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
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
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
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
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
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
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
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": "A"}}`},
			},
			replicationFactor: 5,
			consensus:         false,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase,
			exitCode:          types.TallyExitCodeFilterError,
			filterErr:         types.ErrNoBasicConsensus,
		},
		{
			name:            "MAD filter - One reveal missing without an outlier",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        []bool{false, false, false, false, false},                        // MaxDev = 1*1.5 = 1.5, Median = 5
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": 5}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 6}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 4}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 6}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 5}}`},
			},
			replicationFactor: 6,
			consensus:         true,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			exitCode:          types.TallyExitCodeExecError, // since tally program does not exist
			filterErr:         nil,
		},
		{
			name:            "MAD filter - One outlier",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        []bool{false, false, false, false, true},                         // MaxDev = 1*1.5 = 1.5, Median = 5
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": 5}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 6}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 4}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 6}}`},
				{ExitCode: 0, Reveal: `{"result": {"text": 1}}`},
			},
			replicationFactor: 5,
			consensus:         true,
			consPubKeys:       nil,
			tallyGasUsed:      defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*5,
			exitCode:          types.TallyExitCodeExecError, // since tally program does not exist
			filterErr:         nil,
		},
		{
			name:            "MAD filter - Four reveals missing",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 0, Reveal: `{"result": {"text": 5}}`},
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
			expectedOutliers := make(map[string]bool)
			for i, v := range tt.reveals {
				revealBody := v
				revealBody.Reveal = base64.StdEncoding.EncodeToString([]byte(v.Reveal))
				revealBody.GasUsed = v.GasUsed
				reveals[fmt.Sprintf("%d", i)] = revealBody
				if tt.outliers != nil {
					expectedOutliers[fmt.Sprintf("%d", i)] = tt.outliers[i]
				}
			}

			gasMeter := types.NewGasMeter(1e13, 0, types.DefaultMaxTallyGasLimit, math.NewIntWithDecimal(1, 18), types.DefaultGasCostBase)
			filterRes, tallyRes := f.tallyKeeper.FilterAndTally(f.Context(), types.Request{
				Reveals:           reveals,
				ReplicationFactor: tt.replicationFactor,
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
				PostedGasPrice:    "1000000000000000000", // 1e18
				ExecGasLimit:      100000,
			}, types.DefaultParams(), gasMeter)
			require.NoError(t, err)

			if tt.outliers == nil {
				require.Nil(t, filterRes.Outliers)
			} else {
				for executor, outlier := range expectedOutliers {
					index := slices.Index(filterRes.Executors, executor)
					require.Equal(t, outlier, filterRes.Outliers[index])
				}
			}
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

	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm2(), f.Context().BlockTime())
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
		requestID         string
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
			name:            "Uniform gas reporting (consensus with 1 outlier)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 30000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    60000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(30000),
				"b": math.NewInt(30000),
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
			name:            "Divergent gas reporting (low*2 > median)",
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
			name:            "Divergent gas reporting with multiple lows (low*2 > median)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"lizard":  {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
				"bonobo":  {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"penguin": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
				"zebra":   {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
				"lion":    {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
			},
			requestID:         "646174615F726571756573745F31",
			replicationFactor: 5,
			execGasLimit:      150000,
			expExecGasUsed:    47727 + 25568*4,
			expExecutorGas: map[string]math.Int{
				"lizard":  math.NewInt(25568), // low
				"bonobo":  math.NewInt(25568),
				"penguin": math.NewInt(25568),
				"zebra":   math.NewInt(25568),
				"lion":    math.NewInt(47727), // low
			},
		},
		{
			name:            "Divergent gas reporting with multiple lows, different req ID (low*2 > median)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"lizard":  {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
				"bonobo":  {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"penguin": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
				"zebra":   {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
				"lion":    {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
			},
			requestID:         "646174615F726571756573745F35",
			replicationFactor: 5,
			execGasLimit:      150000,
			expExecGasUsed:    47727 + 25568*4,
			expExecutorGas: map[string]math.Int{
				"lizard":  math.NewInt(47727), // low
				"bonobo":  math.NewInt(25568),
				"penguin": math.NewInt(25568),
				"zebra":   math.NewInt(25568),
				"lion":    math.NewInt(25568), // low
			},
		},
		{
			name:            "Divergent gas reporting with multiple lows (low*2 > median, low is median)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      30000,
			expExecGasUsed:    12000 + 6000*2,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(12000), // low
				"b": math.NewInt(6000),  // low
				"c": math.NewInt(6000),
			},
		},
		{
			name:            "Divergent gas reporting (low*2 < median)",
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
			name:            "Divergent gas reporting with multiple lows (low*2 < median)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 10000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
			},
			replicationFactor: 3,
			execGasLimit:      50000,
			expExecGasUsed:    12000 + 6000*2,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(12000),
				"b": math.NewInt(6000),
				"c": math.NewInt(6000),
			},
		},
		{
			name:            "Divergent gas reporting (low*2 < median)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"zebra":   {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
				"lizard":  {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 20000},
				"bonobo":  {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 35000},
				"penguin": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"lion":    {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
			},
			replicationFactor: 5,
			execGasLimit:      200000,
			expExecGasUsed:    16000 + 28000*4,
			expExecutorGas: map[string]math.Int{
				"lizard":  math.NewInt(28000),
				"bonobo":  math.NewInt(28000),
				"penguin": math.NewInt(16000),
				"zebra":   math.NewInt(28000),
				"lion":    math.NewInt(28000),
			},
		},
		{
			name:            "Divergent gas reporting (lowest_report*2 == median_report)",
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
			name:            "Divergent gas reporting (consensus with 1 outlier, with 2 proxies)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{ // (6000, 18000, 33000) after subtracting proxy gas
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000, ProxyPubKeys: []string{pubKeys[0], pubKeys[1]}},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000, ProxyPubKeys: []string{pubKeys[0], pubKeys[1]}},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000, ProxyPubKeys: []string{pubKeys[0]}},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    42000,
			expExecutorGas: map[string]math.Int{
				"b": math.NewInt(18000),
				"c": math.NewInt(18000),
			},
			expReducedPayout: false,
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
		{
			name:            "MAD uint128 (1 reveal missing, consensus with 2 outliers, uniform gas reporting)",
			tallyInputAsHex: "0200000000002DC6C005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 3, number_type = 0x05, json_path = $.result.text
			reveals: map[string]types.RevealBody{ // median = 400000, MAD = 50000
				"a": {Reveal: `{"result": {"text": 300000, "number": 0}}`, GasUsed: 25000},
				"b": {Reveal: `{"result": {"number": 700000, "number": 0}}`, GasUsed: 25000}, // corrupt
				"c": {Reveal: `{"result": {"text": 400000, "number": 10}}`, GasUsed: 25000},
				"d": {Reveal: `{"result": {"text": 400000, "number": 101}}`, GasUsed: 25000},
				"e": {Reveal: `{"result": {"text": 400000, "number": 0}}`, GasUsed: 25000},
				"f": {Reveal: `{"result": {"text": 340282366920938463463374607431768211456, "number": 0}}`, GasUsed: 25000}, // overflow
				"g": {Reveal: `{"result": {"text": 500000, "number": 0}}`, GasUsed: 25000},
				"h": {Reveal: `{"result": {"text": 500000, "number": 0}}`, GasUsed: 25000},
			},
			replicationFactor: 9,
			execGasLimit:      900000,
			expExecGasUsed:    150000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(25000),
				"c": math.NewInt(25000),
				"d": math.NewInt(25000),
				"e": math.NewInt(25000),
				"g": math.NewInt(25000),
				"h": math.NewInt(25000),
			},
		},
		{
			name:            "MAD uint128 (1 reveal missing, consensus with 2 outliers, divergent gas reporting)",
			tallyInputAsHex: "0200000000002DC6C005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 3, number_type = 0x05, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {Reveal: `{"result": {"text": 300000, "number": 0}}`, GasUsed: 25000},
				"b": {Reveal: `{"result": {"number": 700000, "number": 0}}`, GasUsed: 27000}, // corrupt
				"c": {Reveal: `{"result": {"text": 400000, "number": 10}}`, GasUsed: 21500},
				"d": {Reveal: `{"result": {"text": 400000, "number": 101}}`, GasUsed: 29000},
				"e": {Reveal: `{"result": {"text": 400000, "number": 0}}`, GasUsed: 35000},
				"f": {Reveal: `{"result": {"text": 340282366920938463463374607431768211456, "number": 0}}`, GasUsed: 400}, // overflow
				"g": {Reveal: `{"result": {"text": 500000, "number": 0}}`, GasUsed: 29800},
				"h": {Reveal: `{"result": {"text": 500000, "number": 0}}`, GasUsed: 25000},
			},
			replicationFactor: 9,
			execGasLimit:      900000,
			expExecGasUsed:    156000,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(26000),
				"c": math.NewInt(26000),
				"d": math.NewInt(26000),
				"e": math.NewInt(26000),
				"g": math.NewInt(26000),
				"h": math.NewInt(26000),
			},
		},
		{
			name:            "MAD uint128 (1 reveal missing, consensus with 2 outliers, divergent gas reporting)",
			tallyInputAsHex: "0200000000002DC6C005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 3, number_type = 0x05, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {Reveal: `{"result": {"text": 300000, "number": 0}}`, GasUsed: 25000},
				"b": {Reveal: `{"result": {"number": 700000, "number": 0}}`, GasUsed: 27000}, // corrupt
				"c": {Reveal: `{"result": {"text": 400000, "number": 10}}`, GasUsed: 21500},
				"d": {Reveal: `{"result": {"text": 400000, "number": 101}}`, GasUsed: 29000},
				"e": {Reveal: `{"result": {"text": 400000, "number": 0}}`, GasUsed: 35000},
				"f": {Reveal: `{"result": {"text": 340282366920938463463374607431768211456, "number": 0}}`, GasUsed: 29800}, // overflow
				"g": {Reveal: `{"result": {"text": 500000, "number": 0}}`, GasUsed: 400},
				"h": {Reveal: `{"result": {"text": 500000, "number": 0}}`, GasUsed: 25000},
			},
			replicationFactor: 9,
			execGasLimit:      900000,
			expExecGasUsed:    130800,
			expExecutorGas: map[string]math.Int{
				"a": math.NewInt(26000),
				"c": math.NewInt(26000),
				"d": math.NewInt(26000),
				"e": math.NewInt(26000),
				"g": math.NewInt(800),
				"h": math.NewInt(26000),
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
				PostedGasPrice:    gasPriceStr,
				TallyGasLimit:     types.DefaultMaxTallyGasLimit,
				ExecGasLimit:      tt.execGasLimit,
				TallyProgramID:    hex.EncodeToString(tallyProgram.Hash),
			}
			if tt.requestID != "" {
				request.ID = tt.requestID
			}

			gasMeter := types.NewGasMeter(request.TallyGasLimit, request.ExecGasLimit, types.DefaultMaxTallyGasLimit, gasPrice, types.DefaultGasCostBase)
			_, tallyRes := f.tallyKeeper.FilterAndTally(f.Context(), request, types.DefaultParams(), gasMeter)
			require.NoError(t, err)

			execGasMeter := gasMeter.GetExecutorGasUsed()
			require.Equal(t, len(tt.expExecutorGas), len(execGasMeter))
			for _, exec := range execGasMeter {
				require.Equal(t,
					tt.expExecutorGas[exec.PublicKey].String(),
					exec.Amount.String(),
					fmt.Sprintf("unexpected executor gas for %s", exec.PublicKey),
				)
			}
			require.Equal(t, tt.expReducedPayout, gasMeter.ReducedPayout)
			for _, proxy := range gasMeter.GetProxyGasUsed(request.ID, f.Context().BlockHeight()) {
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
