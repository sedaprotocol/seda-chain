package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
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
		gasUsed           uint64
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
			gasUsed:           defaultParams.FilterGasCostNone,
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
			gasUsed:           defaultParams.FilterGasCostNone,
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
			gasUsed:           defaultParams.FilterGasCostMultiplierMode * 5,
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
			gasUsed:           defaultParams.FilterGasCostMultiplierMode * 5,
			exitCode:          keeper.TallyExitCodeFilterError,
			filterErr:         types.ErrNoBasicConsensus,
		},
		{
			name:              "Mode filter - No reveals",
			tallyInputAsHex:   "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:          []bool{},
			reveals:           map[string]types.RevealBody{},
			replicationFactor: 5,
			consensus:         false,
			consPubKeys:       nil,
			gasUsed:           defaultParams.FilterGasCostMultiplierMode * 5,
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
			gasUsed:           defaultParams.FilterGasCostMultiplierStdDev * 5,
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
			gasUsed:           defaultParams.FilterGasCostMultiplierStdDev * 5,
			exitCode:          keeper.TallyExitCodeFilterError,
			filterErr:         types.ErrNoBasicConsensus,
		},
		{
			name:              "Standard deviation filter - No reveals",
			tallyInputAsHex:   "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:          []bool{},
			reveals:           map[string]types.RevealBody{},
			replicationFactor: 5,
			consensus:         false,
			consPubKeys:       nil,
			gasUsed:           defaultParams.FilterGasCostMultiplierStdDev * 5,
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
				reveals[k] = revealBody
			}

			filterRes, tallyRes, _ := f.tallyKeeper.FilterAndTally(f.Context(), types.Request{
				Reveals:           reveals,
				ReplicationFactor: tt.replicationFactor,
				ConsensusFilter:   base64.StdEncoding.EncodeToString(filterInput),
			}, types.DefaultParams())

			require.Equal(t, tt.outliers, filterRes.Outliers)
			require.Equal(t, tt.gasUsed, filterRes.GasUsed)
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
