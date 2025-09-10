package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigmaMultiplier(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wholeNumber uint64
		fraction    uint64
		bigRat      *big.Rat
		wantErr     bool
	}{
		{
			name:        "valid (0)",
			input:       make([]byte, 8),
			wholeNumber: 0,
			fraction:    0,
			bigRat:      big.NewRat(0, 1),
			wantErr:     false,
		},
		{
			name:        "valid (12.345678)",
			input:       uint64ToBytes(12345678),
			wholeNumber: 12,
			fraction:    345678,
			bigRat:      big.NewRat(12345678, 1e6),
			wantErr:     false,
		},
		{
			name: "valid (max uint64)",
			input: func() []byte { // 18446744073709551615
				b := make([]byte, 8)
				for i := range b {
					b[i] = 0xff
				}
				return b
			}(),
			wholeNumber: 18446744073709,
			fraction:    551615,
			bigRat: func() *big.Rat {
				maxInt := new(big.Int).SetUint64(math.MaxUint64)
				return new(big.Rat).SetFrac(maxInt, big.NewInt(1e6))
			}(),
			wantErr: false,
		},
		{
			name:        "valid (0.000088)",
			input:       uint64ToBytes(88),
			wholeNumber: 0,
			fraction:    88,
			bigRat:      big.NewRat(88, 1e6),
			wantErr:     false,
		},
		{
			name:    "invalid input length",
			input:   []byte{1, 2, 3},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewSigmaMultiplier(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			fmt.Println(sm.WholeNumber())
			fmt.Println(sm.FractionalPart())
			fmt.Println(sm.BigRat())

			require.Equal(t, tt.wholeNumber, sm.WholeNumber())
			require.Equal(t, tt.fraction, sm.FractionalPart())
			require.Equal(t, tt.bigRat, sm.BigRat())
		})
	}
}

// Helper function to convert uint64 to big-endian bytes
func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}
