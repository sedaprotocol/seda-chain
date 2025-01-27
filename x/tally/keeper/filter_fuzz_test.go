package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func FuzzStdDevFilter(f *testing.F) {
	fixture := initFixture(f)

	err := fixture.tallyKeeper.SetParams(fixture.Context(), types.DefaultParams())
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, n0, n1, n2, n3, n4, n5, n6, n7, n8, n9 uint64) {
		nums := []uint64{n0, n1, n2, n3, n4, n5, n6, n7, n8, n9}
		var reveals []types.RevealBody
		for _, num := range nums {
			revealStr := fmt.Sprintf(`{"result": {"text": %d}}`, num)
			reveals = append(reveals, types.RevealBody{
				Reveal: base64.StdEncoding.EncodeToString([]byte(revealStr)),
			})
		}

		// Prepare inputs and execute filter.
		bz := make([]byte, 8)
		filterHex := fmt.Sprintf("02%s05000000000000000b726573756C742E74657874", hex.EncodeToString(bz)) // max_sigma = neighborDist, number_type = int64, json_path = result.text
		filterInput, err := hex.DecodeString(filterHex)
		require.NoError(t, err)

		gasMeter := types.NewGasMeter(1e13, 0, types.DefaultMaxTallyGasLimit, sdkmath.NewIntWithDecimal(1, 18), types.DefaultGasCostBase)

		_, _ = keeper.ExecuteFilter(
			reveals,
			base64.StdEncoding.EncodeToString(filterInput),
			uint16(len(reveals)),
			types.DefaultParams(),
			gasMeter,
		)
	})
}
