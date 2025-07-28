package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFullDataRequestFlow(t *testing.T) {
	f := initFixture(t)

	tests := []struct {
		name              string
		memo              string
		replicationFactor int
		numCommits        int
		numReveals        int
		timeout           bool
		expExitCode       uint32
	}{
		{
			name:              "full single commit-reveal",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo0")),
			replicationFactor: 1,
			numCommits:        1,
			numReveals:        1,
			timeout:           false,
			expExitCode:       0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyPubKeys := []string{"03b27f2df0cbdb5cdadff5b4be0c9fda5aa3a59557ef6d0b49b4298ef42c8ce2b0"}
			err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(bondDenom, math.NewInt(1000000000000000000)))
			require.NoError(t, err)

			drID, stakers := f.commitRevealDataRequest(
				t, tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
				commitRevealConfig{
					requestHeight: 1,
					requestMemo:   tt.memo,
					reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
					proxyPubKeys:  proxyPubKeys,
					gasUsed:       150000000000000000,
				})

			beforeBalance := f.bankKeeper.GetBalance(f.Context(), stakers[0].address, bondDenom)
			posterBeforeBalance := f.bankKeeper.GetBalance(f.Context(), f.coreAuthority, bondDenom)

			err = f.tallyKeeper.EndBlock(f.Context())
			require.NoError(t, err)
			require.NotContains(t, f.logBuf.String(), "ERR")

			// TODO query get_staker pending_withdrawal and check diff
			// Verify the staker did not pay for the transactions
			afterBalance := f.bankKeeper.GetBalance(f.Context(), stakers[0].address, bondDenom)
			diff := afterBalance.Sub(beforeBalance)
			require.Equal(t, "0aseda", diff.String())

			// Verify the poster paid for execution
			afterPostBalance := f.bankKeeper.GetBalance(f.Context(), f.coreAuthority, bondDenom)
			diff = afterPostBalance.Sub(posterBeforeBalance)
			require.NotEqual(t, "0aseda", diff.String(), "Poster should have payed for execution")

			dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, tt.expExitCode, dataResult.ExitCode)

			dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
			require.NoError(t, err)
			require.Contains(t, dataResults, *dataResult)

			f.coreKeeper.Allowlist.Walk(f.Context(), nil, func(key string) (stop bool, err error) {
				fmt.Println("allowlist")
				fmt.Println(key)
				return false, nil
			})
		})
	}
}
