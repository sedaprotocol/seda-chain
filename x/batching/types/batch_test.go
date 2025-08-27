package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

// TestComputeBatchID tests batch ID computation based on a sample batch
// created from the EVM contracts repository:
//
//		BatchNumber: 1
//		BlockHeight: 0
//		ValidatorsRoot: 0xbb9dcde527b03bcf4022f44c13ffce6eb445e765cc3064dd5c6093ca781a7ebc
//		DataResultsRoot: 0xb361e88bec2624b327671f15761d4c24915dbc11503343f3e325d0d97061552a
//		ProvingMetadata: 0x0000000000000000000000000000000000000000000000000000000000000000
//	 ID: 0xafd4483e5cfa9954ecc23c78da0abd2c1a92d0e8be6b3e5bf554cd0ea5ac305b
func TestComputeBatchID(t *testing.T) {
	valRoot, err := hex.DecodeString("bb9dcde527b03bcf4022f44c13ffce6eb445e765cc3064dd5c6093ca781a7ebc")
	require.NoError(t, err)
	dataRoot, err := hex.DecodeString("b361e88bec2624b327671f15761d4c24915dbc11503343f3e325d0d97061552a")
	require.NoError(t, err)
	expectedBatchID, err := hex.DecodeString("afd4483e5cfa9954ecc23c78da0abd2c1a92d0e8be6b3e5bf554cd0ea5ac305b")
	require.NoError(t, err)

	batchID := types.ComputeBatchID(1, 0, valRoot, dataRoot, make([]byte, 32))
	require.Equal(t, expectedBatchID, batchID)
}
