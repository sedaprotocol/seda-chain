package types

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDataResult_TryHash(t *testing.T) {
	gasUsed := math.NewInt(1000)
	tests := []struct {
		name         string
		dr           DataResult
		want         string // Expected hash
		wantErr      bool
		expectedHash string
	}{
		{
			name: "valid data result",
			dr: DataResult{
				Version:        "1.0",
				DrId:           "1234567890abcdef",
				Consensus:      true,
				ExitCode:       0,
				Result:         []byte("test result"),
				BlockHeight:    1000,
				BlockTimestamp: 1234567890,
				GasUsed:        &gasUsed,
				PaybackAddress: "YWRkcmVzczEyMzQ=", // base64 encoded "address1234"
				SedaPayload:    "cGF5bG9hZDEyMzQ=", // base64 encoded "payload1234"
			},
			wantErr:      false,
			expectedHash: "fe2027b5fd3862a32c675f358461ffc8264dc03d3e2d4525fd525d209d4603a8",
		},
		{
			name: "invalid dr_id",
			dr: DataResult{
				Version: "1.0",
				DrId:    "invalid hex", // This will cause an error
			},
			wantErr: true,
		},
		{
			name: "invalid payback address",
			dr: DataResult{
				Version:        "1.0",
				DrId:           "1234567890abcdef",
				PaybackAddress: "invalid base64!@#$",
			},
			wantErr: true,
		},
		{
			name: "invalid seda payload",
			dr: DataResult{
				Version:     "1.0",
				DrId:        "1234567890abcdef",
				SedaPayload: "invalid base64!@#$",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dr.TryHash()

			// Check error cases
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Equal(t, tt.expectedHash, got)
			}

			// For the valid case, we also want to ensure the hash is consistent
			if !tt.wantErr {
				got2, err := tt.dr.TryHash()
				if err != nil {
					t.Errorf("Second TryHash() call failed: %v", err)
				}
				if got != got2 {
					t.Error("TryHash() is not deterministic")
				}
			}
		})
	}
}
