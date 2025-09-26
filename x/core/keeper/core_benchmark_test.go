package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// BenchmarkDataRequestFlow benchmarks core end blockers with 1000 tally programs
// randomly selected from test wasm catalog.
func BenchmarkDataRequestFlow(b *testing.B) {
	f := testutil.InitFixture(b)
	f.AddStakers(b, 5)

	proxyPubKeys := []string{"03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0"}
	err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(testutil.BondDenom, math.NewInt(1000000000000000000)))
	require.NoError(b, err)

	params := types.DefaultParams()
	params.TallyConfig.MaxTalliesPerBlock = 1000
	f.CoreKeeper.SetParams(f.Context(), params)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 1000; j++ {
			dr := testutil.NewTestDRWithRandomPrograms(
				[]byte("reveal"),
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
				150000000000000000,
				0,
				proxyPubKeys,
				1,
				f.Context().BlockTime(),
			)
			dr.ExecuteDataRequestFlow(f, 1, 1, false)
		}

		b.StartTimer()
		err = f.CoreKeeper.EndBlock(f.Context())
		require.NoError(b, err)

		f.AddBlock()
	}
}

// BenchmarkBigTallyPrograms benchmarks core end blockers with 1000 executions
// of a big tally program (slightly less than 1MB in size).
func BenchmarkBigTallyPrograms(b *testing.B) {
	f := testutil.InitFixture(b)
	f.AddStakers(b, 5)

	params := types.DefaultParams()
	params.TallyConfig.MaxTalliesPerBlock = 1000
	f.CoreKeeper.SetParams(f.Context(), params)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.BigWasm(), f.Context().BlockTime())

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 1000; j++ {
			dr := testutil.NewTestDR(
				execProgram.Hash, execProgram.Hash,
				[]byte("reveal"),
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
				150000000000000000,
				0,
				[]string{},
				1,
			)
			dr.ExecuteDataRequestFlow(f, 1, 1, false)
		}

		b.StartTimer()
		err := f.CoreKeeper.EndBlock(f.Context())
		require.NoError(b, err)

		f.AddBlock()
	}
}
