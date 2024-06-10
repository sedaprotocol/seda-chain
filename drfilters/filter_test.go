package drfilters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutliers_None(t *testing.T) {
	tests := []struct {
		name       string
		tallyInput []byte
		want       []bool
		reveals    map[string]any
		wantErr    error
	}{
		{
			name:       "None filter",
			tallyInput: []byte{193, 128}, // tallyProp{ Algo: 0}
			want:       []bool{true, true, true, true, true},
			reveals: map[string]any{
				"this":   1,
				"reveal": 2,
				"has":    3,
				"five":   4,
				"values": 5,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Outliers(tt.tallyInput, tt.reveals)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.Equal(t, got, tt.want)
		})
	}
}
