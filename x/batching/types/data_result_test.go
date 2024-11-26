package types_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"

// 	"github.com/sedaprotocol/seda-chain/x/batching/types"
// )

// // Expected RESULT ID for the following Data Result:
// // {
// //     "version": "0.0.1",
// //     "dr_id": "74d7e8c9a77b7b4777153a32fcdf2424489f24cd59d3043eb2a30be7bba48306",
// //     "consensus": true,
// //     "exit_code": 0,
// //     "result": "Ghkvq84TmIuEmU1ClubNxBjVXi8df5QhiNQEC5T8V6w=",
// //     "block_height": 12345,
// //     "gas_used": 20,
// //     "payback_address": "",
// //     "seda_payload": ""
// //   }
// // let expected_result_id = "f6fc1b4ea295b00537bbe3e918793699c43dbc924ee7df650da567a095238150";

// func Test_DataResult_TryHash(t *testing.T) {
// 	// Create a sample DataResult
// 	dr := &types.DataResult{
// 		Version:        "0.0.1",
// 		DrId:           "74d7e8c9a77b7b4777153a32fcdf2424489f24cd59d3043eb2a30be7bba48306",
// 		Consensus:      true,
// 		ExitCode:       0,
// 		Result:         []byte("Ghkvq84TmIuEmU1ClubNxBjVXi8df5QhiNQEC5T8V6w="),
// 		BlockHeight:    12345,
// 		GasUsed:        20,
// 		PaybackAddress: "",
// 		SedaPayload:    "",
// 	}

// 	// Call TryHash
// 	hash, err := dr.TryHash()
// 	require.NoError(t, err)

// 	require.Equal(t, "f6fc1b4ea295b00537bbe3e918793699c43dbc924ee7df650da567a095238150", hash)

// 	// // Assert hash is not empty
// 	// assert.NotEmpty(t, hash)

// 	// // Assert hash is a valid hex string of length 64 (32 bytes)
// 	// assert.Len(t, hash, 64)
// 	// assert.Regexp(t, "^[0-9a-f]{64}$", hash)

// 	// // Test consistency
// 	// hash2, err := dr.TryHash()
// 	// assert.NoError(t, err)
// 	// assert.Equal(t, hash, hash2)

// 	// // Test with different data
// 	// dr.Result = []byte("different result")
// 	// differentHash, err := dr.TryHash()
// 	// assert.NoError(t, err)
// 	// assert.NotEqual(t, hash, differentHash)

// 	// // Test error case: invalid DrId
// 	// dr.DrId = "invalid"
// 	// _, err = dr.TryHash()
// 	// assert.Error(t, err)

// 	// // Test error case: invalid PaybackAddress
// 	// dr.DrId = "0123456789abcdef" // Reset to valid DrId
// 	// dr.PaybackAddress = "invalid base64"
// 	// _, err = dr.TryHash()
// 	// assert.Error(t, err)

// 	// // Test error case: invalid SedaPayload
// 	// dr.PaybackAddress = base64.StdEncoding.EncodeToString([]byte("payback_address")) // Reset to valid PaybackAddress
// 	// dr.SedaPayload = "invalid base64"
// 	// _, err = dr.TryHash()
// 	// assert.Error(t, err)
// }
