package utils_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/app/utils"
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
			expected: "bccb08ea97abfa6350c70db4207620796408060f72e5019c3acbe855c59568b2",
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

func TestRootFromLeaves(t *testing.T) {
	tests := []struct {
		name     string
		entries  []string // hex-encoded leaves
		expected string
	}{
		{
			name: "ten leaves",
			entries: []string{
				"38519bc19a6c21b2e4d5c07f5c317a04907d74428e84ce53d7e31660363697e3",
				"efaaa7508118895a0d04b782da3824d32e8068bc149eb9cb64d5035bfdc23d75",
				"08acd70e8df30e046db9b1f428e801b98426493443630c1a10ed41ca831b59dc",
				"b83b5ecaf184b6525c995e98885c3c26105981d368fba6a32f118d963a8b2f0e",
				"5ea122e40dd0c50ea173ed35af483f3391d910051d0b81b5764962bc7259c75a",
				"1b36b415df57e691bc923f10a6dc481caf1b31a77f3514c443e78e07a9692e70",
				"15c3bea022ed633d5aeabfeba41955d19f23ad533c95d50b121b1cb0c585c2f7",
				"e2f6ad28acf622aad072b4af09ff3e4588c312461fd9f34f87a30398c4874c6f",
				"66931389acd9b85bf84eb1b4b391e0100affed572a4b5695bcbb07c1655fbfe7",
				"49f4e02c70dc0cff353603c5af12911e3fe408e8263005cbe2877e2677789469",
			},
			expected: "2c5073e9c4308e65eb22152f62611c6f604d6b427c6514bbd1effaf282d9614a",
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
			got := utils.RootFromLeaves(entries)
			if !bytes.Equal(got, expected) {
				t.Errorf("RootFromLeaves() = %x, want %x", got, expected)
			}
		})
	}
}
