package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
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
		filterGasUsed     uint64
		exitCode          int
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
			filterGasUsed:     defaultParams.FilterGasCostNone,
			exitCode:          keeper.TallyExitCodeExecError, // since tally program does not exist
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
			filterGasUsed:     0,
			exitCode:          keeper.TallyExitCodeFilterError,
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
			filterGasUsed:     defaultParams.FilterGasCostMultiplierMode * 5,
			exitCode:          keeper.TallyExitCodeExecError, // since tally program does not exist
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
			filterGasUsed:     0,
			exitCode:          keeper.TallyExitCodeFilterError,
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
			filterGasUsed:     defaultParams.FilterGasCostMultiplierStdDev * 5,
			exitCode:          keeper.TallyExitCodeExecError, // since tally program does not exist
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
			filterGasUsed:     0,
			exitCode:          keeper.TallyExitCodeFilterError,
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

			filterRes, tallyRes, _ := f.tallyKeeper.FilterAndTally(f.Context(), types.Request{
				Reveals:           reveals,
				ReplicationFactor: tt.replicationFactor,
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
				GasPrice:          "1000000000000000000", // 1e18
				ExecGasLimit:      100000,
			}, types.DefaultParams(), math.NewInt(1000000000000000000))
			require.NoError(t, err)

			require.Equal(t, tt.outliers, filterRes.Outliers)
			require.Equal(t, tt.filterGasUsed, filterRes.GasUsed)
			require.Equal(t, tt.consensus, filterRes.Consensus)
			require.Equal(t, tt.consensus, tallyRes.Consensus)
			require.Equal(t, tt.exitCode, tallyRes.ExitInfo.ExitCode)
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

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm2(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	tests := []struct {
		name               string
		tallyInputAsHex    string
		reveals            map[string]types.RevealBody
		replicationFactor  uint16
		execGasLimit       uint64
		expExecGasUsed     uint64
		expExecutorRewards map[string]math.Int
		expProxyRewards    map[string]math.Int
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
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(30000, 18),
				"b": math.NewIntWithDecimal(30000, 18),
				"c": math.NewIntWithDecimal(30000, 18),
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
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(20000, 18),
				"b": math.NewIntWithDecimal(20000, 18),
				"c": math.NewIntWithDecimal(20000, 18),
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
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(30000, 18),
				"b": math.NewIntWithDecimal(30000, 18),
				"c": math.NewIntWithDecimal(30000, 18),
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
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(30000, 18),
				"b": math.NewIntWithDecimal(30000, 18),
				"c": math.NewIntWithDecimal(30000, 18),
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
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(16000, 18),
				"b": math.NewIntWithDecimal(16000, 18),
				"c": math.NewIntWithDecimal(16000, 18),
			},
		},
		{
			name:            "Divergent gas reporting (1)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 28000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 30000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 32000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    90000,
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(43448, 18),
				"b": math.NewIntWithDecimal(23275, 18),
				"c": math.NewIntWithDecimal(23275, 18),
			},
		},
		{
			name:            "Divergent gas reporting (2)",
			tallyInputAsHex: "00",
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 20000},
				"c": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 35000},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    56000,
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(16000, 18),
				"b": math.NewIntWithDecimal(20000, 18),
				"c": math.NewIntWithDecimal(20000, 18),
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
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(16000*0.8, 18),
				"b": math.NewIntWithDecimal(20000*0.8, 18),
				"c": math.NewIntWithDecimal(20000*0.8, 18),
			},
		},
		{
			name:            "Divergent gas reporting (mode no consensus)",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // mode, json_path = $.result.text
			reveals: map[string]types.RevealBody{
				"a": {ExitCode: 0, Reveal: `{"result": {"text": "A"}}`, GasUsed: 8000, ProxyPubKeys: []string{"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8"}},
				"b": {ExitCode: 0, Reveal: `{"result": {"text": "B"}}`, GasUsed: 20000, ProxyPubKeys: []string{"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8"}},
				"c": {ExitCode: 1, Reveal: `{"result": {"text": "B"}}`, GasUsed: 35000, ProxyPubKeys: []string{"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8"}},
			},
			replicationFactor: 3,
			execGasLimit:      90000,
			expExecGasUsed:    52000, // (7000, 19000, 34000) after subtracting proxy gas
			expExecutorRewards: map[string]math.Int{
				"a": math.NewIntWithDecimal(14000*0.8, 18),
				"b": math.NewIntWithDecimal(19000*0.8, 18),
				"c": math.NewIntWithDecimal(19000*0.8, 18),
			},
			expProxyRewards: map[string]math.Int{
				"161b0d3a1efbf2f7d2f130f68a2ccf8f8f3220e8": math.NewIntWithDecimal(1000, 18), // = proxyFee / gasPrice
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
								Fee: &proxyFee,
							},
						)
						require.NoError(t, err)
					}
				}
			}

			gasPriceStr := "1000000000000000000" // 1e18
			gasPrice, ok := math.NewIntFromString(gasPriceStr)
			require.True(t, ok)

			_, tallyRes, payoutRecord := f.tallyKeeper.FilterAndTally(
				f.Context(),
				types.Request{
					Reveals:           reveals,
					ReplicationFactor: tt.replicationFactor,
					ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
					GasPrice:          gasPriceStr,
					ExecGasLimit:      tt.execGasLimit,
					TallyProgramID:    hex.EncodeToString(tallyProgram.Hash),
					// TODO tally gas limit
				}, types.DefaultParams(), gasPrice)
			require.NoError(t, err)

			for _, distMsg := range payoutRecord.ExecDists {
				require.Equal(t,
					tt.expExecutorRewards[distMsg.ExecutorReward.Identity].String(),
					distMsg.ExecutorReward.Amount.String(),
				)
			}
			for _, distMsg := range payoutRecord.ProxyDists {
				require.Equal(t,
					tt.expProxyRewards[hex.EncodeToString(distMsg.DataProxyReward.To)].String(),
					distMsg.DataProxyReward.Amount.String(),
				)
			}
			require.Equal(t, tt.expExecGasUsed, tallyRes.ExecGasUsed)
		})
	}
}
