package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func TestEndBlock(t *testing.T) {
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

			drID := f.executeDataRequestFlow(
				t, nil, nil,
				tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
				commitRevealConfig{
					requestHeight: 1,
					requestMemo:   tt.memo,
					reveal:        []byte("reveal"),
					proxyPubKeys:  proxyPubKeys,
					gasUsed:       150000000000000000,
				})

			// Data request should be in the tallying status.
			dr, err := f.coreKeeper.GetDataRequest(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, dr.Status)
			for executor := range dr.Reveals {
				_, err := f.coreKeeper.GetRevealBody(f.Context(), drID, executor)
				require.NoError(t, err)
			}

			// TODO re-activate more detailed checks
			// beforeBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			// posterBeforeBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)

			err = f.coreKeeper.EndBlock(f.Context())
			require.NoError(t, err)

			// Data request should have been removed from the store.
			dr, err = f.coreKeeper.GetDataRequest(f.Context(), drID)
			require.Error(t, err)
			for executor := range dr.Reveals {
				_, err := f.coreKeeper.GetRevealBody(f.Context(), drID, executor)
				require.Error(t, err)
			}

			// // TODO query get_staker pending_withdrawal and check diff
			// // Verify the staker did not pay for the transactions
			// afterBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			// diff := afterBalance.Sub(beforeBalance)
			// require.Equal(t, "0aseda", diff.String())

			// // Verify the poster paid for execution
			// afterPostBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)
			// diff = afterPostBalance.Sub(posterBeforeBalance)
			// require.NotEqual(t, "0aseda", diff.String(), "Poster should have paid for execution")

			// dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			// require.NoError(t, err)
			// // TODO map oracle program to exit code
			// // require.Equal(t, tt.expExitCode, dataResult.ExitCode)

			// dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
			// require.NoError(t, err)
			// require.Contains(t, dataResults, *dataResult)
		})
	}
}
