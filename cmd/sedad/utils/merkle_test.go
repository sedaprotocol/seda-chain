package utils_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/sedaprotocol/seda-chain/cmd/sedad/utils"
	"github.com/stretchr/testify/require"
)

func TestRootFromEntries(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string // hex-encoded entries
		expected string
	}{
		{
			name:     "empty entries",
			entries:  []string{},
			expected: "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		},
		{
			name:     "single entry",
			entries:  []string{"a6f4b3e2d9c18f75ba7e3c21d5f8a239ce4b7a2d3f85c9e6b7d2f3a4c9d8f2"},
			expected: "10580d33d16c58046698b4ac4ef92bf7d73fcd778cab0885e95d8e8e1e24ad5f",
		},
		{
			name: "two entries",
			entries: []string{
				"22e922a540bb7af9b2456ecb0663619804db6e736f28702398c594b93e526acaefa2d0fb95f0c45c",
				"a54cfd0002b948adba2acd65025cad55aa8558ccd84e421f6e6b9519f2e465a577be05ba32de0317",
			},
			expected: "fbb5a63d65fec2bf6f5903994c763f354f66b9a45711b8342979a9f67ef8e27b",
		},
		{
			name: "five entries",
			entries: []string{
				"5ef99dd9c6c8070b95a8b2e6f709f457cbba95d4883b719f5ba4ef6fc81cfe119bf9df2e4c49ea06",
				"4e9bbe0628ddee1a89578919948d513d1abd6ff2fedb400cefa5c3e6ba3ff912919c5d103fcafaad",
				"9c09ebebf888a7cac4dbadadf15f6c07d38b4ea1d3faac2d74e9aed2654274471bd4498df4da23cb",
				"3940ae8fd7928039a8d630301e25d9ede15bd94700f4b73c5543b3089bc47efcfdaa0a4104418dd6",
				"67b17cf31449cf9c738738b4382afa789d8bb794eab3a42ac0c6990d67d23c8f7e785b2d3d317fda",
			},
			expected: "5784dd7a32bb4df072e7f1053d3ae4eb5d8833ecb52062e95f4a26ec9b2d244b",
		},
	}

	for _, tt := range tests {
		expected, err := hex.DecodeString(tt.expected)
		require.NoError(t, err)
		entries := make([][]byte, len(tt.entries))
		for i := range tt.entries {
			entries[i], err = hex.DecodeString(tt.entries[i])
			require.NoError(t, err)
		}
		t.Run(tt.name, func(t *testing.T) {
			got := utils.RootFromEntries(entries)
			if !bytes.Equal(got, expected) {
				t.Errorf("RootFromEntries() = %x, want %x", got, expected)
			}
		})
	}
}
