package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name            string
		tallyInputAsHex string
		outliers        []int
		reveals         []types.RevealBody
		consensus       bool
		wantErr         error
	}{
		{
			name:            "None filter",
			tallyInputAsHex: "00",
			outliers:        []int{0, 0, 0, 0, 0},
			reveals: []types.RevealBody{
				{},
				{},
				{},
				{},
				{},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Happy Path",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 1, 0, 1, 0, 0},
			reveals: []types.RevealBody{
				{Reveal: `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`},
				{Reveal: `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`},
				{Reveal: `{"but":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{Reveal: `{"matter":"ignore this", "result": {"text": "A", "number": 10}}`},
				{Reveal: `{"matter":"ignore this", "result": {"text": "A", "number": 10}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - One outlier but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "B", "number": 101}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Multiple modes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0, 0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A"}}`},
				{Reveal: `{"result": {"text": "A"}}`},
				{Reveal: `{"result": {"text": "A"}}`},
				{Reveal: `{"result": {"text": "B"}}`},
				{Reveal: `{"result": {"text": "B"}}`},
				{Reveal: `{"result": {"text": "B"}}`},
				{Reveal: `{"result": {"text": "C"}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - One corrupt reveal but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 1, 0},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"resultt": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 101}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - No consensus on exit code",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0, 0, 0},
			reveals: []types.RevealBody{
				{ExitCode: 1, Reveal: `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`},
				{ExitCode: 1, Reveal: `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`},
				{ExitCode: 1, Reveal: `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`},
				{ExitCode: 0, Reveal: `{"but":"ignore this", "result": {"text": "B", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"matter":"ignore this", "result": {"text": "C", "number": 10}}`},
			},
			consensus: false,
			wantErr:   types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - Corrupt due to too many bad exit codes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0, 0, 0},
			reveals: []types.RevealBody{
				{ExitCode: 1, Reveal: `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`},
				{ExitCode: 1, Reveal: `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`},
				{ExitCode: 1, Reveal: `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`},
				{ExitCode: 1, Reveal: `{"but":"ignore this", "result": {"text": "B", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{ExitCode: 0, Reveal: `{"matter":"ignore this", "result": {"text": "C", "number": 10}}`},
			},
			consensus: false,
			wantErr:   types.ErrCorruptReveals,
		},
		{
			name:            "Mode filter - Uniform reveals",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0, 0, 0},
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
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Basic consensus but corrupt due to too many bad exit codes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0, 0, 0},
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
			consensus: false,
			wantErr:   types.ErrCorruptReveals,
		},
		{
			name:            "Mode filter with proxy pubkeys - No basic consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0, 0, 0},
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
			consensus: false,
			wantErr:   types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - Half with different reveals but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 1, 0},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"invalid_proxy_pubkey"}, Reveal: `{"result": {"text": "mac"}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - No consensus due to non-zero exit code invalidating data",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 1, 1},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"invalid_proxy_pubkey"}, Reveal: `{"result": {"text": "mac"}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - No consensus with exit code invalidating a reveal",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 1},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": ""}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - One reports bad pubkey but is not an outlier",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{1, 0, 0, 0},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"qwerty"}, Reveal: `{"result": {"text": "windows"}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Too many bad exit codes",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{0, 0, 0, 0},
			reveals: []types.RevealBody{
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "mac"}}`},
				{ExitCode: 0, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
				{ExitCode: 1, ProxyPubKeys: []string{"02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3"}, Reveal: `{"result": {"text": "windows"}}`},
			},
			consensus: false,
			wantErr:   types.ErrNoBasicConsensus,
		},
		{
			name:            "Mode filter - Bad exit code but consensus",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{1, 0, 0, 1, 0, 0, 0},
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
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Consensus not reached due to exit code",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{1, 0, 0, 1, 1, 0},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`, ExitCode: 1},
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "B", "number": 10}}`},
				{Reveal: `{"result": {"text": "C", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Consensus not reached due to corrupt reveal",
			tallyInputAsHex: "01000000000000000D242E726573756C742E74657874", // json_path = $.result.text
			outliers:        []int{1, 0, 0, 1, 1, 0},
			reveals: []types.RevealBody{
				{Reveal: `{"resalt": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 101}}`},
				{Reveal: `{"result": {"text": "B", "number": 10}}`},
				{Reveal: `{"result": {"text": "C", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter uint64",
			tallyInputAsHex: "02000000000016E36003000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = uint64, json_path = $.result.text
			outliers:        []int{1, 0, 0, 0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 10}}`},
				{Reveal: `{"result": {"text": 6, "number": 101}}`},
				{Reveal: `{"result": {"text": 7, "number": 0}}`},
				{Reveal: `{"result": {"text": 8, "number": 0}}`},
				{Reveal: `{"result": {"text": 9, "number": 0}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter int64",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        []int{1, 0, 0, 0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 10}}`},
				{Reveal: `{"result": {"text": 6, "number": 101}}`},
				{Reveal: `{"result": {"text": 7, "number": 0}}`},
				{Reveal: `{"result": {"text": 8, "number": 0}}`},
				{Reveal: `{"result": {"text": 9, "number": 0}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter - Empty reveal",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = uint64, json_path = $.result.text
			outliers:        []int{},
			reveals:         []types.RevealBody{},
			consensus:       false,
			wantErr:         types.ErrEmptyReveals,
		},
		{
			name:            "Standard deviation filter - Single reveal",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = uint64, json_path = $.result.text
			outliers:        []int{0},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter - One corrupt reveal",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = uint64, json_path = $.result.text
			outliers:        []int{1, 0, 0, 0, 1, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 10}}`},
				{Reveal: `{"result": {"text": 6, "number": 101}}`},
				{Reveal: `{"result": {"text": 7, "number": 0}}`},
				{Reveal: `{"result": {"number": 0}}`}, // corrupt
				{Reveal: `{"result": {"text": 9, "number": 0}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter - Max sigma 1.55",
			tallyInputAsHex: "02000000000017A6B003000000000000000D242E726573756C742E74657874", // max_sigma = 1.55, number_type = uint64, json_path = $.result.text
			outliers:        []int{1, 0, 0, 0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 10}}`},
				{Reveal: `{"result": {"text": 6, "number": 101}}`},
				{Reveal: `{"result": {"text": 7, "number": 0}}`},
				{Reveal: `{"result": {"text": 8, "number": 0}}`},
				{Reveal: `{"result": {"text": 9, "number": 0}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter - Max sigma 1.45",
			tallyInputAsHex: "02000000000016201003000000000000000D242E726573756C742E74657874", // max_sigma = 1.45, number_type = uint64, json_path = $.result.text
			outliers:        []int{1, 1, 0, 0, 1, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 4, "number": 0}}`},
				{Reveal: `{"result": {"text": 5, "number": 10}}`},
				{Reveal: `{"result": {"text": 6, "number": 101}}`},
				{Reveal: `{"result": {"text": 7, "number": 0}}`},
				{Reveal: `{"result": {"text": 8, "number": 0}}`},
				{Reveal: `{"result": {"text": 9, "number": 0}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter int64 with negative reveals",
			tallyInputAsHex: "02000000000016E36001000000000000000D242E726573756C742E74657874", // max_sigma = 1.5, number_type = int64, json_path = $.result.text
			outliers:        []int{1, 0, 0, 0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": -4, "number": 0}}`},
				{Reveal: `{"result": {"text": -5, "number": 10}}`},
				{Reveal: `{"result": {"text": -6, "number": 101}}`},
				{Reveal: `{"result": {"text": -7, "number": 0}}`},
				{Reveal: `{"result": {"text": -8, "number": 0}}`},
				{Reveal: `{"result": {"text": -9, "number": 0}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Standard deviation filter int64 median -0.5",
			tallyInputAsHex: "02000000000007A12001000000000000000D242E726573756C742E74657874", // max_sigma = 0.5, number_type = int64, json_path = $.result.text
			outliers:        []int{1, 0, 0, 1},
			reveals: []types.RevealBody{
				{Reveal: `{"result": {"text": 1, "number": 0}}`},
				{Reveal: `{"result": {"text": 0, "number": 0}}`},
				{Reveal: `{"result": {"text": -1, "number": 10}}`},
				{Reveal: `{"result": {"text": -2, "number": 10}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := hex.DecodeString(tt.tallyInputAsHex)
			require.NoError(t, err)

			// For illustration
			for i := 0; i < len(tt.reveals); i++ {
				tt.reveals[i].Reveal = base64.StdEncoding.EncodeToString([]byte(tt.reveals[i].Reveal))
			}

			// Since ApplyFilter assumes the pubkeys are sorted.
			for i := range tt.reveals {
				sort.Strings(tt.reveals[i].ProxyPubKeys)
			}

			outliers, cons, err := keeper.ApplyFilter(filter, tt.reveals)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.outliers, outliers)
			require.Equal(t, tt.consensus, cons)
		})
	}
}
