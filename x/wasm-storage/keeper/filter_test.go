package keeper_test

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/stretchr/testify/require"
)

func TestFilterNone(t *testing.T) {
	tests := []struct {
		name       string
		tallyInput []byte
		want       []int
		reveals    []keeper.RevealBody
		wantErr    error
	}{
		{
			name:       "None filter",
			tallyInput: []byte{0, 26, 129, 196}, // filterProp{ Algo: 0}
			want:       []int{0, 0, 0, 0, 0},
			reveals: []keeper.RevealBody{
				{},
				{},
				{},
				{},
				{},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := keeper.ApplyFilter(tt.tallyInput, tt.reveals)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
