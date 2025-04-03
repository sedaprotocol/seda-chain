package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func TestFilter(t *testing.T) {
	f := initFixture(t)

	defaultParams := types.DefaultParams()
	err := f.tallyKeeper.SetParams(f.Context(), defaultParams)
	require.NoError(t, err)

	tests := []struct {
		name            string
		tallyInputAsHex string
		outliers        []bool
		reveals         []types.RevealBody
		consensus       bool
		consPubKeys     []string // expected proxy public keys in basic consensus
		tallyGasUsed    uint64
		wantErr         error
	}{
		{
			name:            "None filter",
			tallyInputAsHex: "00",
			outliers:        make([]bool, 5),
			reveals: []types.RevealBody{
				{},
				{},
				{},
				{},
				{},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostNone,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - Happy Path",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, false, true, false, true, false, false},
			reveals: []types.RevealBody{
				{Reveal: `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`},
				{Reveal: `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`},
				{Reveal: `{"but":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{Reveal: `{"matter":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"matter":"ignore this", "result": {"text": "A", "number": 10}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*7,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - One outlier but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, false, true},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "B", "number": 101}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*3,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - exactly 2/3rd reports an error",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, false, true},
			reveals: []types.RevealBody{
				{ExitCode: 1, Reveal: ``},
				{ExitCode: 2, Reveal: ``},
				{ExitCode: 0, Reveal: `{"result": {"text": "A", "number": 0}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*3,
			wantErr:      types.ErrConsensusInError,
		},
		{
			name:            "Mode filter - Multiple modes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A"}}`},
				{Reveal: `{"result": {"text": "A"}}`},
				{Reveal: `{"result": {"text": "A"}}`},
				{Reveal: `{"result": {"text": "B"}}`},
				{Reveal: `{"result": {"text": "B"}}`},
				{Reveal: `{"result": {"text": "B"}}`},
				{Reveal: `{"result": {"text": "C"}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*7,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "Mode filter - One corrupt reveal but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, true, false},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"resultt": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 101}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*3,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - No consensus on exit code",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 1, Reveal: `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`},
				{ExitCode: 1, Reveal: `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`},
				{ExitCode: 1, Reveal: `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`},
				{ExitCode: 0, Reveal: `{"but":"ignore this", "result": {"text": "B", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"matter":"ignore this", "result": {"text": "C", "number": 10}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + 0,
			wantErr:      types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - >2/3 bad exit codes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, false, false, false, true, false},
			reveals: []types.RevealBody{
				{ExitCode: 1, Reveal: `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`},
				{ExitCode: 1, Reveal: `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`},
				{ExitCode: 1, Reveal: `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`},
				{ExitCode: 1, Reveal: `{"but":"ignore this", "result": {"text": "B", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{ExitCode: 1, Reveal: `{"matter":"ignore this", "result": {"text": "C", "number": 10}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*6,
			wantErr:      types.ErrConsensusInError,
		},
		{
			name:            "Mode filter - Uniform reveals",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        make([]bool, 6),
			reveals: []types.RevealBody{
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
			},
			consensus: true,
			consPubKeys: []string{
				"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
				"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
				"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
			},
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*6,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - >2/3 bad exit codes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, true, false, false, false, false},
			reveals: []types.RevealBody{
				{
					ExitCode: 1,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 1,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 1,
					ProxyPubKeys: []string{
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 1,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 1,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
			},
			consensus: true,
			consPubKeys: []string{
				"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
				"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
				"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
			},
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*6,
			wantErr:      types.ErrConsensusInError,
		},
		{
			name:            "Mode filter with proxy pubkeys - No basic consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{
					ExitCode: 1,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode:     0,
					ProxyPubKeys: []string{},
					Reveal:       `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
				{
					ExitCode: 0,
					ProxyPubKeys: []string{
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
						"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
						"034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
					},
					Reveal: `{"result": {"text": "A"}}`,
				},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + 0,
			wantErr:      types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - Half with different reveals but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{false, false, true, false},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"invalid_proxy_pubkey"}, Reveal: `{"result": {"text": "mac"}}`},
			},
			consensus:    true,
			consPubKeys:  []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"},
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*4,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - No consensus due to non-zero exit code invalidating data",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"invalid_proxy_pubkey"}, Reveal: `{"result": {"text": "mac"}}`},
			},
			consensus:    false,
			consPubKeys:  []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"},
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*4,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "Mode filter - No consensus with exit code invalidating a reveal",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": ""}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
			},
			consensus:    false,
			consPubKeys:  []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"},
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*4,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "Mode filter - One reports bad pubkey but is not an outlier",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{true, false, false, false},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"qwerty"}, Reveal: `{"result": {"text": "windows"}}`},
			},
			consensus:    true,
			consPubKeys:  []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"},
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*4,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - Too many bad exit codes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + 0,
			wantErr:      types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - Bad exit code but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []bool{true, false, false, true, false, false, false},
			reveals: []types.RevealBody{
				{
					ExitCode: 1,
					Reveal:   `{"xx":"ignore this", "result": {"text": "A", "number": 0}}`,
				},
				{Reveal: `{"xx":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"xx":"ignore this", "result": {"text": "A", "number": 101}}`},
				{Reveal: `{"xx":"ignore this", "result": {"text": "B", "number": 10}}`},
				{Reveal: `{"xx":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"xx":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"xx":"ignore this", "result": {"text": "A", "number": 10}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*7,
			wantErr:      nil,
		},
		{
			name:            "Mode filter - Consensus not reached due to exit code",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`, ExitCode: 1},
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "B", "number": 10}}`},
				{Reveal: `{"result": {"text": "C", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*6,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "Mode filter - Consensus not reached due to corrupt reveal",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"resalt": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 101}}`},
				{Reveal: `{"result": {"text": "B", "number": 10}}`},
				{Reveal: `{"result": {"text": "C", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*6,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD - All zeros",
			tallyInputAsHex: "02000000000016E36006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.5, number_type = 0x00, json_path = $.result.text
			outliers:        []bool{false, false, false, false},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 0}}`},
				{Reveal: `{"result": {"text": 0}}`},
				{Reveal: `{"result": {"text": 0}}`},
				{Reveal: `{"result": {"text": 0}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*4,
			wantErr:      nil,
		},
		{
			name:            "MAD - No consensus despite big outlier attack",
			tallyInputAsHex: "02000000000016E36006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.5, number_type = 0x00, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 1000}}`},
				{Reveal: `{"result": {"text": 1000}}`},
				{Reveal: `{"result": {"text": 115792089237316195423570985008687907853269}}`},
				{Reveal: `{"result": {"text": 5000}}`},
				{Reveal: `{"result": {"text": 5500}}`},
				{Reveal: `{"result": {"text": 5400}}`},
				{Reveal: `{"result": {"text": 5300}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*7,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD int32 (sigma_multiplier = 1.0)",
			tallyInputAsHex: "0200000000000F424005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x01, json_path = $.result.text
			outliers:        []bool{true, false, false, true, false, false},
			reveals: []types.RevealBody{ // median = 6.5, MAD = 1.5, max_dev = 1.5
				{Reveal: `{"result": {"text": 4}}`}, // outlier
				{Reveal: `{"result": {"text": 5}}`},
				{Reveal: `{"result": {"text": 6}}`},
				{Reveal: `{"result": {"text": 9}}`}, // outlier
				{Reveal: `{"result": {"text": 7}}`},
				{Reveal: `{"result": {"text": 8}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      nil,
		},
		{
			name:            "MAD uint32 (Some invalid reveals)",
			tallyInputAsHex: "0200000000000F424005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x01, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4294967295}}`},        // ok (max of uint64)
				{Reveal: `{"result": {"text": 4294967296}}`},        // overflow
				{Reveal: `{"result": {"text": 4294967295}}`},        // ok
				{Reveal: `{"result": {"text": -100, "number": 0}}`}, // negative
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*4,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD uint64 (Some invalid reveals)",
			tallyInputAsHex: "0200000000000F424005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x03, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 18446744073709551615}}`}, // ok (max of uint64)
				{Reveal: `{"result": {"text": 18446744073709551616}}`}, // overflow
				{Reveal: `{"result": {"text": 18446744073709551615}}`}, // ok
				{Reveal: `{"result": {"text": -100, "number": 0}}`},    // negative
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*4,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD uint128 (Some invalid reveals)",
			tallyInputAsHex: "0200000000000F424005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x05, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 340282366920938463463374607431768211455}}`}, // ok (max of uint128)
				{Reveal: `{"result": {"text": 340282366920938463463374607431768211456}}`}, // overflow
				{Reveal: `{"result": {"text": 340282366920938463463374607431768211455}}`}, // ok
				{Reveal: `{"result": {"text": -100, "number": 0}}`},                       // negative
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*4,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD int64 (With an overflow)",
			tallyInputAsHex: "0200000000001E848002000000000000000D242E726573756C742E74657874", // sigma_multiplier = 2.0, number_type = 0x02, json_path = $.result.text
			outliers:        []bool{false, false, false, false, true, true},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4}}`},
				{Reveal: `{"result": {"text": 5}}`},
				{Reveal: `{"result": {"text": 6}}`},
				{Reveal: `{"result": {"text": 7}}`},
				{Reveal: `{"result": {"text": -9223372036854775809}}`}, // overflow
				{Reveal: `{"result": {"text": 9}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      nil,
		},
		{
			name:            "MAD (Single reveal)",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.5, number_type = 0x01, json_path = $.result.text
			outliers:        []bool{false},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD,
			wantErr:      nil,
		},
		{
			name:            "MAD int32 (One overflow)",
			tallyInputAsHex: "0200000000001E848000000000000000000D242E726573756C742E74657874", // sigma_multiplier = 2.0, number_type = 0x00, json_path = $.result.text
			outliers:        []bool{false, false, false, false, false, true},
			reveals: []types.RevealBody{ // median = 6, MAD = 1, max_dev = 2 => [4,8]
				{Reveal: `{"result": {"text": 4}}`},
				{Reveal: `{"result": {"text": 5}}`},
				{Reveal: `{"result": {"text": 6}}`},
				{Reveal: `{"result": {"text": 7}}`},
				{Reveal: `{"result": {"text": 6}}`},
				{Reveal: `{"result": {"text": 2147483648}}`}, // overflow
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      nil,
		},
		{
			name:            "MAD int32 (Negative numbers with sigma_multiplier = 2.0)",
			tallyInputAsHex: "0200000000001E848000000000000000000D242E726573756C742E74657874", // sigma_multiplier = 2.0, number_type = 0x00, json_path = $.result.text
			outliers:        []bool{false, false, false, false, false, false},
			reveals: []types.RevealBody{ // median = -6.5, MAD = 1.5, max_dev = 3 => [-9.5, -3.5]
				{Reveal: `{"result": {"text": -4, "number": 0}}`},
				{Reveal: `{"result": {"text": -5, "number": 10}}`},
				{Reveal: `{"result": {"text": -6, "number": 101}}`},
				{Reveal: `{"result": {"text": -7, "number": 0}}`},
				{Reveal: `{"result": {"text": -8, "number": 0}}`},
				{Reveal: `{"result": {"text": -9, "number": 0}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      nil,
		},
		{
			name:            "MAD int32 (Negative numbers with sigma_multiplier = 1.0)",
			tallyInputAsHex: "0200000000000F424000000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x00, json_path = $.result.text
			outliers:        []bool{true, false, false, false, false, true},
			reveals: []types.RevealBody{ // median = -6.5, MAD = 1.5, max_dev = 1.5 => [-8, -5]
				{Reveal: `{"result": {"text": -4, "number": 0}}`}, // outlier
				{Reveal: `{"result": {"text": -5, "number": 10}}`},
				{Reveal: `{"result": {"text": -6, "number": 101}}`},
				{Reveal: `{"result": {"text": -7, "number": 0}}`},
				{Reveal: `{"result": {"text": -8, "number": 0}}`},
				{Reveal: `{"result": {"text": -9, "number": 0}}`}, // outlier
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      nil,
		},
		{
			name:            "MAD uint128 (One corrupt and one overflow)",
			tallyInputAsHex: "0200000000002DC6C005000000000000000D242E726573756C742E74657874", // sigma_multiplier = 3, number_type = 0x05, json_path = $.result.text
			outliers:        []bool{false, true, false, false, false, true, false, false},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 300000, "number": 0}}`},
				{Reveal: `{"result": {"number": 700000, "number": 0}}`}, // corrupt
				{Reveal: `{"result": {"text": 400000, "number": 10}}`},
				{Reveal: `{"result": {"text": 400000, "number": 101}}`},
				{Reveal: `{"result": {"text": 400000, "number": 0}}`},
				{Reveal: `{"result": {"text": 340282366920938463463374607431768211456, "number": 0}}`}, // overflow
				{Reveal: `{"result": {"text": 500000, "number": 0}}`},
				{Reveal: `{"result": {"text": 500000, "number": 0}}`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*8,
			wantErr:      nil,
		},
		{
			name:            "MAD int256",
			tallyInputAsHex: "02000000000016E36003000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.5, number_type = 0x06, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{ // median = 4.5, MAD = 0.5, max_dev = 0.75 => [3.75, 5.25]
				{Reveal: `{"result": {"text": 2, "number": 0}}`}, // outlier
				{Reveal: `{"result": {"text": 4, "number": 10}}`},
				{Reveal: `{"result": {"text": 4, "number": 101}}`},
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 0}}`},
				{Reveal: `{"result": {"text": 7, "number": 0}}`}, // outlier
				{Reveal: `{"result": {"text": 9, "number": 0}}`}, // outlier
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*8,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD int256 (Negative numbers)",
			tallyInputAsHex: "0200000000000F424006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x06, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{ // median = -28679, MAD = 404.5 => [-29083.5, -28274.5]
				{Reveal: `{"result": {"text": -28930, "number": 0}}`},
				{Reveal: `{"result": {"text": -28000, "number": 10}}`},  // outlier
				{Reveal: `{"result": {"text": -30005, "number": 101}}`}, // outlier
				{Reveal: `{"result": {"text": -28600, "number": 0}}`},
				{Reveal: `{"result": {"text": -28758, "number": 0}}`},
				{Reveal: `{"result": {"text": -28121, "number": 0}}`}, // outlier
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD int256 (Negative numbers (2))",
			tallyInputAsHex: "0200000000000F424006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x06, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": -28930, "number": 0}}`},
				{Reveal: `{"result": {"text": -28000, "number": 10}}`},
				{Reveal: `{"result": {"text": -29005, "number": 101}}`},
				{Reveal: `{"result": {"text": -28600, "number": 0}}`},
				{Reveal: `{"result": {"text": -27758, "number": 0}}`},
				{Reveal: `{"result": {"text": -28121, "number": 0}}`},
			},
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD int256 (Negative numbers (3))",
			tallyInputAsHex: "0200000000000F8C7806000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.019, number_type = 0x06, json_path = $.result.text
			outliers:        nil,
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": -28930, "number": 0}}`}, // outlier
				{Reveal: `{"result": {"text": -28000, "number": 10}}`},
				{Reveal: `{"result": {"text": -29005, "number": 101}}`}, // outlier
				{Reveal: `{"result": {"text": -28600, "number": 0}}`},
				{Reveal: `{"result": {"text": -27758, "number": 0}}`}, // outlier
				{Reveal: `{"result": {"text": -28121, "number": 0}}`},
			}, // median = -28360.5, MAD = 464, range = [-28833.316, -27887.684]
			consensus:    false,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      types.ErrNoConsensus,
		},
		{
			name:            "MAD int256 (Negative numbers (4))",
			tallyInputAsHex: "02000000000013D62006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.3, number_type = 0x06, json_path = $.result.text
			outliers:        []bool{false, false, true, false, false, false},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": -28930, "number": 0}}`},
				{Reveal: `{"result": {"text": -28000, "number": 10}}`},
				{Reveal: `{"result": {"text": -29005, "number": 101}}`}, // outlier
				{Reveal: `{"result": {"text": -28600, "number": 0}}`},
				{Reveal: `{"result": {"text": -27758, "number": 0}}`},
				{Reveal: `{"result": {"text": -28121, "number": 0}}`},
			}, // median = -28360.5, MAD = 465, max_dev = 604.5, range = [-28965, -27756]
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*6,
			wantErr:      nil,
		},
		{
			name:            "MAD int256 (Very large numbers)",
			tallyInputAsHex: "0200000000000F424006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x06, json_path = $.result.text
			outliers:        []bool{true, false, false, false, false, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 3000000000000000000000000000000000000000, "number": 0}}`}, // outlier
				{Reveal: `{"result": {"text": 4000000000000000000000000000000000000000, "number": 10}}`},
				{Reveal: `{"result": {"text": 4000000000000000000000000000000000000000, "number": 101}}`},
				{Reveal: `{"result": {"text": 4000000000000000000000000000000000000000, "number": 0}}`},
				{Reveal: `{"result": {"text": 5000000000000000000000000000000000000000, "number": 0}}`},
				{Reveal: `{"result": {"text": 5000000000000000000000000000000000000000, "number": 0}}`},
				{Reveal: `{"result": {"text": 5000000000000000000000000000000000000000, "number": 0}}`},
				{Reveal: `{"result": {"text": 6000000000000000000000000000000000000000, "number": 0}}`}, // outlier
			}, // median = 4.5, MAD = 0.5, range = [4, 5]
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*8,
			wantErr:      nil,
		},
		{
			name:            "MAD int256 (Some reveals too large)",
			tallyInputAsHex: "0200000000000F424006000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x06, json_path = $.result.text
			outliers:        []bool{true, false, false, false, false, true, false, false, false},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819969, "number": 0}}`},   // too small
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 10}}`},  // ok (min of int256)
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 101}}`}, // ok
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 0}}`},   // ok
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 0}}`},   // ok
				{Reveal: `{"result": {"text": 115792089237316195423570985008687907853269984665640564039457584007913129639936, "number": 0}}`},   // too large (max of uint256 + 1)
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 0}}`},   // ok
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 0}}`},   // ok
				{Reveal: `{"result": {"text": -57896044618658097711785492504343953926634992332820282019728792003956564819968, "number": 0}}`},   // ok
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*9,
			wantErr:      nil,
		},
		{
			name:            "MAD uint256 (Some reveals negative)",
			tallyInputAsHex: "0200000000000F424007000000000000000D242E726573756C742E74657874", // sigma_multiplier = 1.0, number_type = 0x07, json_path = $.result.text
			outliers:        []bool{false, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 115792089237316195423570985008687907853269984665640564039457584007913129639935}}`}, // ok (max of uint256)
				{Reveal: `{"result": {"text": 115792089237316195423570985008687907853269984665640564039457584007913129639935}}`}, // ok
				{Reveal: `{"result": {"text": 115792089237316195423570985008687907853269984665640564039457584007913129639935}}`}, // ok
				{Reveal: `{"result": {"text": -100, "number": 0}}`},                                                              // negative
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*4,
			wantErr:      nil,
		},
		{
			name:            "Std dev filter (JSON value number)",
			tallyInputAsHex: "02000000000016E36000000000000000000124", // sigma_multiplier = 1.5, number_type = 0x00, json_path = $
			outliers:        []bool{false, false, true, false},
			reveals: []types.RevealBody{
				{Reveal: `3136`},
				{Reveal: `3136`},
				{Reveal: `"3136"`}, // string, not number
				{Reveal: `3136`},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*4,
			wantErr:      nil,
		},
		{
			name:            "Mode filter (JSON value string)",
			tallyInputAsHex: "01000000000000000124", // json_path = $
			outliers:        []bool{false, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: `"yes"`},
				{Reveal: `"yes"`},
				{Reveal: `"yes"`},
				{Reveal: `yes`}, // invalid due to no surrounding double quotes
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*4,
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterInput, err := hex.DecodeString(tt.tallyInputAsHex)
			require.NoError(t, err)

			for i := 0; i < len(tt.reveals); i++ {
				tt.reveals[i].Reveal = base64.StdEncoding.EncodeToString([]byte(tt.reveals[i].Reveal))
			}

			reveals := make([]types.Reveal, len(tt.reveals))
			for i := range tt.reveals {
				sort.Strings(tt.reveals[i].ProxyPubKeys)
				reveals[i] = types.Reveal{
					Executor:   fmt.Sprintf("%d", i),
					RevealBody: tt.reveals[i],
				}
			}

			gasMeter := types.NewGasMeter(1e13, 0, types.DefaultMaxTallyGasLimit, math.NewIntWithDecimal(1, 18), types.DefaultGasCostBase)

			result, err := keeper.ExecuteFilter(
				reveals,
				base64.StdEncoding.EncodeToString(filterInput), uint16(len(tt.reveals)),
				types.DefaultParams(),
				gasMeter,
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.consPubKeys == nil {
				require.Nil(t, nil, result.ProxyPubKeys)
			} else {
				for _, pk := range tt.consPubKeys {
					require.Contains(t, result.ProxyPubKeys, pk)
				}
			}

			require.Equal(t, tt.outliers, result.Outliers)
			require.Equal(t, tt.consensus, result.Consensus)
			require.Equal(t, tt.tallyGasUsed, gasMeter.TallyGasUsed())
		})
	}
}

// TestFilterWildcard tests filters with JSON paths containing wildcard expressions.

func TestFilterWildcard(t *testing.T) {
	f := initFixture(t)

	defaultParams := types.DefaultParams()
	err := f.tallyKeeper.SetParams(f.Context(), defaultParams)
	require.NoError(t, err)

	tests := []struct {
		name           string
		filterInputHex string // filter input without JSON path as hex
		jsonPath       string // JSON path, if applicable (separated to accomodate odd length byte slices)
		outliers       []bool
		reveals        []types.RevealBody
		consensus      bool
		consPubKeys    []string // expected proxy public keys in basic consensus
		tallyGasUsed   uint64
		wantErr        error
	}{
		{
			name:           "Mode filter - Nested elements",
			filterInputHex: "010000000000000014",
			jsonPath:       "$.store.*[*].country",
			outliers:       []bool{false, false, false, true, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: sampleRevealNested},
				{Reveal: sampleRevealNested},
				{Reveal: sampleRevealNested2},
				{},
				{Reveal: sampleRevealNested},
				{Reveal: sampleRevealNested2},
				{},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*7,
			wantErr:      nil,
		},
		{
			name:           "Mode filter",
			filterInputHex: "010000000000000012",
			jsonPath:       "$.store.*[*].title",
			outliers:       []bool{false, false, false, true, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: sampleReveal},
				{Reveal: sampleReveal},
				{Reveal: sampleReveal2},
				{},
				{Reveal: sampleReveal},
				{Reveal: sampleReveal2},
				{},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*7,
			wantErr:      nil,
		},
		{
			name:           "Mode filter - Prices differ but the lowest is the same",
			filterInputHex: "010000000000000012",
			jsonPath:       "$.store.*[*].price",
			outliers:       []bool{false, false, false, true, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: sampleReveal},
				{Reveal: sampleReveal},
				{Reveal: sampleReveal2},
				{Reveal: sampleRevealNested},
				{Reveal: sampleReveal},
				{Reveal: sampleReveal2},
				{Reveal: sampleRevealNested},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMode*7,
			wantErr:      nil,
		},
		{
			name:           "MAD",
			filterInputHex: "0200000000000F4240070000000000000012", // sigma_multiplier = 1.0, number_type = 0x07
			jsonPath:       "$.store.*[*].price",
			outliers:       []bool{false, false, false, true, false, false, true},
			reveals: []types.RevealBody{
				{Reveal: sampleReveal},
				{Reveal: sampleReveal},
				{Reveal: sampleReveal2},
				{},
				{Reveal: sampleReveal},
				{Reveal: sampleReveal2},
				{},
			},
			consensus:    true,
			consPubKeys:  nil,
			tallyGasUsed: defaultParams.GasCostBase + defaultParams.FilterGasCostMultiplierMAD*7,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterInput, err := hex.DecodeString(tt.filterInputHex)
			require.NoError(t, err)
			filterInput = append(filterInput, []byte(tt.jsonPath)...)

			for i := 0; i < len(tt.reveals); i++ {
				tt.reveals[i].Reveal = base64.StdEncoding.EncodeToString([]byte(tt.reveals[i].Reveal))
			}

			reveals := make([]types.Reveal, len(tt.reveals))
			for i := range tt.reveals {
				sort.Strings(tt.reveals[i].ProxyPubKeys)
				reveals[i] = types.Reveal{
					Executor:   fmt.Sprintf("%d", i),
					RevealBody: tt.reveals[i],
				}
			}

			gasMeter := types.NewGasMeter(1e13, 0, types.DefaultMaxTallyGasLimit, math.NewIntWithDecimal(1, 18), types.DefaultGasCostBase)

			result, err := keeper.ExecuteFilter(
				reveals,
				base64.StdEncoding.EncodeToString(filterInput), uint16(len(tt.reveals)),
				types.DefaultParams(),
				gasMeter,
			)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.consPubKeys == nil {
				require.Nil(t, nil, result.ProxyPubKeys)
			} else {
				for _, pk := range tt.consPubKeys {
					require.Contains(t, result.ProxyPubKeys, pk)
				}
			}

			require.Equal(t, tt.outliers, result.Outliers)
			require.Equal(t, tt.consensus, result.Consensus)
			require.Equal(t, tt.tallyGasUsed, gasMeter.TallyGasUsed())
		})
	}
}

var sampleReveal = `{
  "store": {
    "book": [
      {
        "category": "fiction",
        "author": "J.K. Rowling",
        "title": "Harry Potter and the Sorcerer's Stone",
        "price": 1999
      },
      {
        "category": "fantasy",
        "author": "J.R.R. Tolkien",
        "title": "The Hobbit",
        "price": 1499
      },
      {
        "category": "science fiction",
        "author": "Isaac Asimov",
        "title": "Foundation",
        "price": 1899
      }
    ]
  }
}`

var sampleReveal2 = `{
	"store": {
	  "book": [
		{
		  "category": "science fiction",
		  "author": "Isaac Asimov",
		  "title": "Foundation",
		  "price": 5899
		},
		{
		  "category": "fiction",
		  "author": "J.K. Rowling",
		  "title": "Harry Potter and the Sorcerer's Stone",
		  "price": 3999
		},
		{
		  "category": "fantasy",
		  "author": "J.R.R. Tolkien",
		  "title": "The Hobbit",
		  "price": 1499
		}
	  ]
	}
}`

var sampleRevealNested = `{
  "store": {
    "books": [
      {
        "category": "fiction",
        "author": "Author 1",
        "title": "Book Title 1",
        "country": [
          {
            "name": "angola"
          },
          {
            "name": "brazil"
          }
        ]
      },
      {
        "category": "non-fiction",
        "author": "Author 2",
        "title": "Book Title 2",
        "country": [
          {
            "name": "angola"
          },
          {
            "name": "brazil"
          }
        ]
      },
      {
        "category": "fiction",
        "author": "Author 3",
        "title": "Book Title 3",
        "country": [
          {
            "name": "angola"
          },
          {
            "name": "brazil"
          }
        ]
      }
    ],
    "music": [
      {
        "category": "pop",
        "country": "China"
      },
      {
        "category": "rock",
        "country": "Solomon Islands"
      }
    ]
  }
}`

var sampleRevealNested2 = `{
	"store": {
	  "music": [
		{
			"category": "rock",
			"country": "Solomon Islands"
		},
		{
			"category": "pop",
			"country": "China"
		}
	  ],
	  "books": [
	  	{
		  "category": "fiction",
		  "author": "Author 3",
		  "title": "Book Title 3",
		  "country": [
			{
			  "name": "angola"
			},
			{
			  "name": "brazil"
			}
		  ]
		},
		{
		  "category": "fiction",
		  "author": "Author 1",
		  "title": "Book Title 1",
		  "country": [
			{
			  "name": "angola"
			},
			{
			  "name": "brazil"
			}
		  ]
		},
		{
		  "category": "non-fiction",
		  "author": "Author 2",
		  "title": "Book Title 2",
		  "country": [
			{
			  "name": "angola"
			},
			{
			  "name": "brazil"
			}
		  ]
		}
	  ]
	}
}`
