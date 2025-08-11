package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkDataRequestFlow(b *testing.B) {
	b.StopTimer()

	f := initFixture(b)

	proxyPubKeys := []string{"03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0"}
	err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(bondDenom, math.NewInt(1000000000000000000)))
	require.NoError(b, err)

	tt := struct {
		name string
		// memo              string
		replicationFactor int
		numCommits        int
		numReveals        int
		timeout           bool
		expExitCode       uint32
	}{
		name: "full single commit-reveal",
		// memo:              base64.StdEncoding.EncodeToString([]byte("memo0")),
		replicationFactor: 1,
		numCommits:        1,
		numReveals:        1,
		timeout:           false,
		expExitCode:       0,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.commitRevealDataRequest(
			b, tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
			commitRevealConfig{
				requestHeight: 1,
				requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
				reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
				proxyPubKeys:  proxyPubKeys,
				gasUsed:       150000000000000000,
			})
		err = f.keeper.EndBlock(f.Context())
		require.NoError(b, err)
		err = f.batchingKeeper.EndBlock(f.Context())
		require.NoError(b, err)
		f.AddBlock()
	}

	/*
		b.Run(tt.name, func(b *testing.B) {
			f.commitRevealDataRequest(
				b, tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
				commitRevealConfig{
					requestHeight: 1,
					requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
					reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
					proxyPubKeys:  proxyPubKeys,
					gasUsed:       150000000000000000,
				})

			// Core Endblock
			err = f.keeper.EndBlock(f.Context())
			require.NoError(b, err)
			// require.NotContains(b, f.logBuf.String(), "ERR")

			// dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			// require.NoError(b, err)
			// require.Equal(b, tt.expExitCode, dataResult.ExitCode)

			// dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
			// require.NoError(b, err)
			// require.Contains(b, dataResults, *dataResult)

			// Batching Endblock
			err = f.batchingKeeper.EndBlock(f.Context())
			require.NoError(b, err)

			f.AddBlock()
			fmt.Println("iteration ", b.N)
			// batches, err := f.batchingKeeper.GetAllBatches(f.Context())
			// require.NoError(b, err)
			// require.Equal(b, 1, len(batches))
		})
	*/
}
