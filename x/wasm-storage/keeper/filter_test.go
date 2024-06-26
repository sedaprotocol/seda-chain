package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/stretchr/testify/require"
)

func TestFilterNone(t *testing.T) {
	tests := []struct {
		name            string
		tallyInputAsHex string
		outliers        []int
		reveals         []keeper.RevealBody
		consensus       bool
		wantErr         error
	}{
		{
			name:            "None filter",
			tallyInputAsHex: "00", // filterProp{ Algo: 0}
			outliers:        []int{0, 0, 0, 0, 0},
			reveals: []keeper.RevealBody{
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
			tallyInputAsHex: "01000000000000000b726573756C742E74657874", // json_path = result.text
			outliers:        []int{0, 0, 1, 0, 1, 0, 0},
			reveals: []keeper.RevealBody{
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
			name:            "Mode filter - At least 2/3 satisfies consensus",
			tallyInputAsHex: "01000000000000000b726573756C742E74657874", // json_path = result.text
			outliers:        []int{0, 0, 1},
			reveals: []keeper.RevealBody{
				{Reveal: `{"result": {"text": "A", "number": 0}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "B", "number": 101}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Consensus due to non exit code",
			tallyInputAsHex: "01000000000000000b726573756C742E74657874", // json_path = result.text
			outliers:        []int{1, 1, 1, 1, 1, 1},
			reveals: []keeper.RevealBody{
				{
					ExitCode: 1,
					Reveal:   `{"high_level_prop1":"ignore this", "result": {"text": "A", "number": 0}}`,
				},
				{
					ExitCode: 1,
					Reveal:   `{"makes_this_json":"ignore this", "result": {"text": "A", "number": 10}}`,
				},
				{
					ExitCode: 1,
					Reveal:   `{"unstructured":"ignore this", "result": {"text": "B", "number": 101}}`,
				},
				{Reveal: `{"but":"ignore this", "result": {"text": "B", "number": 10}}`},
				{Reveal: `{"it_does_not":"ignore this", "result": {"text": "C", "number": 10}}`},
				{Reveal: `{"matter":"ignore this", "result": {"text": "C", "number": 10}}`},
			},
			consensus: true,
			wantErr:   nil,
		},
		{
			name:            "Mode filter - Valid reveal marked outlier due to non exit code [still consensus]",
			tallyInputAsHex: "01000000000000000b726573756C742E74657874", // json_path = result.text
			outliers:        []int{1, 0, 0, 1, 0, 0, 0},
			reveals: []keeper.RevealBody{
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
			tallyInputAsHex: "01000000000000000b726573756C742E74657874", // json_path = result.text
			outliers:        []int{1, 1, 1, 1, 1, 1},
			reveals: []keeper.RevealBody{
				{
					ExitCode: 1,
					Reveal:   `{"result": {"text": "A", "number": 0}}`,
				},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 101}}`},
				{Reveal: `{"result": {"text": "B", "number": 10}}`},
				{Reveal: `{"result": {"text": "C", "number": 10}}`},
				{Reveal: `{"result": {"text": "A", "number": 10}}`},
			},
			consensus: false,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := hex.DecodeString(tt.tallyInputAsHex)
			require.NoError(t, err)
			outliers, cons, err := keeper.ApplyFilter(filter, tt.reveals)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.outliers, outliers)
			require.Equal(t, tt.consensus, cons)
		})
	}
}
