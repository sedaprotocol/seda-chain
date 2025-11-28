package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// bytes -> strings -> bytes
func TestDataRequestIndexBytes(t *testing.T) {
	indexBytes := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 208, 255, 255, 255, 255, 255, 255, 255, 255, 150, 167, 194, 102, 61, 95, 189, 2, 223, 91, 45, 180, 76, 194, 203, 131, 220, 191, 76, 239, 249, 70, 189, 29, 33, 30, 227, 180, 10, 184, 166, 204}
	index := DataRequestIndex(indexBytes)
	reconstructedIndex, err := DataRequestIndexFromStrings(index.Strings())
	require.NoError(t, err)
	require.True(t, bytes.Equal(index, reconstructedIndex))
	require.Equal(t, index.Strings(), reconstructedIndex.Strings())
}

// strings -> bytes -> strings
func TestDataRequestIndexStrings(t *testing.T) {
	indexStrs := []string{
		"1400000000000000000", // posted gas price
		"6166411",             // height
		"49ecda48cfbf58a65c413db373b38d1952906de47c102c6ad7bf3e32e1fe2764", // data request ID
	}
	index, err := DataRequestIndexFromStrings(indexStrs)
	require.NoError(t, err)
	require.Equal(t, indexStrs, index.Strings())
}
