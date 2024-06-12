package keeper_test

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/stretchr/testify/require"
)

func TestOutliers_None(t *testing.T) {
	tests := []struct {
		name       string
		tallyInput []byte
		want       []bool
		reveals    []keeper.RevealBody
		wantErr    error
	}{
		{
			name:       "None filter",
			tallyInput: []byte{193, 128}, // filterProp{ Algo: 0}
			want:       []bool{false, false, false, false, false},
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
			got, _, err := keeper.Outliers(tt.tallyInput, tt.reveals)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.Equal(t, got, tt.want)
		})
	}
}
