package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// CheckStakerIndexing verifies that the staker public key - index mapping.
// If exists is set to false, it checks that the mapping does not exist.
func (f Fixture) CheckStakerIndexing(t *testing.T, pubKey string, index uint32, exists bool) {
	t.Helper()

	gotIndex, err := f.CoreKeeper.GetStakerIndex(f.Context(), pubKey)
	if exists {
		require.NoError(t, err)
		require.Equal(t, index, gotIndex)
	} else {
		require.Error(t, err)
	}

	gotPubKey, err := f.CoreKeeper.GetStakerKey(f.Context(), index)
	if exists {
		require.NoError(t, err)
		require.Equal(t, pubKey, gotPubKey)
	} else {
		require.Error(t, err)
	}
}
