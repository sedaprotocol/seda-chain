package keeper_test

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func FuzzStdDevFilter(f *testing.F) {
	f.Fuzz(func(t *testing.T, n0, n1, n2, n3, n4, n5, n6, n7, n8, n9 int64) {
		source := rand.NewSource(time.Now().UnixNano())
		t.Log("random testing with seed", source.Int63())

		nums := []int64{n0, n1, n2, n3, n4, n5, n6, n7, n8, n9}
		var reveals []types.RevealBody
		for _, num := range nums {
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(num))
			revealStr := fmt.Sprintf(`{"result": {"text": "%s"}}`, base64.StdEncoding.EncodeToString(buf))
			reveals = append(reveals, types.RevealBody{
				Reveal: base64.StdEncoding.EncodeToString([]byte(revealStr)),
			})
		}

		// Compute stats about nums using float arithmetics.
		length := len(nums)
		numsSorted := make([]int64, length)
		copy(numsSorted, nums)
		slices.Sort(numsSorted)
		var median float64
		mid := length / 2
		if length%2 == 1 {
			median = float64(numsSorted[mid]+numsSorted[mid]) / 2
		} else {
			median = float64(numsSorted[mid-1]+numsSorted[mid]) / 2
		}
		neighborDist := float64(numsSorted[mid] - numsSorted[mid-1])
		expOutliers := make([]int, len(nums))
		for i, num := range nums {
			if math.Abs(float64(num)-median) > neighborDist {
				expOutliers[i] = 1
			}
		}

		bz := make([]byte, 8)
		binary.BigEndian.PutUint64(bz, uint64(neighborDist*1e6))
		filterHex := fmt.Sprintf("02%s01000000000000000b726573756C742E74657874", hex.EncodeToString(bz)) // max_sigma = neighborDist, number_type = int64, json_path = result.text
		filter, err := hex.DecodeString(filterHex)
		require.NoError(t, err)

		outliers, _, err := keeper.ApplyFilter(filter, reveals)
		require.Equal(t, expOutliers, outliers)
		require.ErrorIs(t, err, nil)
	})
}
