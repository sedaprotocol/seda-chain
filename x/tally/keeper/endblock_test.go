package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
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
		{
			name:              "full 5 commit-reveals",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo1")),
			replicationFactor: 5,
			numCommits:        5,
			numReveals:        5,
			timeout:           false,
			expExitCode:       0,
		},
		{
			name:              "commit timeout",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo2")),
			replicationFactor: 2,
			numCommits:        0,
			numReveals:        0,
			timeout:           true,
			expExitCode:       types.TallyExitCodeNotEnoughCommits,
		},
		{
			name:              "commit timeout with 1 commit",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo3")),
			replicationFactor: 2,
			numCommits:        1,
			numReveals:        0,
			timeout:           true,
			expExitCode:       types.TallyExitCodeNotEnoughCommits,
		},
		{
			name:              "commit timeout with 2 commits",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo4")),
			replicationFactor: 2,
			numCommits:        1,
			numReveals:        0,
			timeout:           true,
			expExitCode:       types.TallyExitCodeNotEnoughCommits,
		},
		{
			name:              "reveal timeout with no reveals",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo5")),
			replicationFactor: 2,
			numCommits:        2,
			numReveals:        0,
			timeout:           true,
			expExitCode:       types.TallyExitCodeFilterError,
		},
		{
			name:              "reveal timeout with 2 reveals",
			memo:              base64.StdEncoding.EncodeToString([]byte("memo6")),
			replicationFactor: 3,
			numCommits:        3,
			numReveals:        2,
			timeout:           true,
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
					requestMemo:  tt.memo,
					reveal:       base64.StdEncoding.EncodeToString([]byte("reveal")),
					proxyPubKeys: proxyPubKeys,
					gasUsed:      150000000000000000,
				})

			beforeBalance := f.bankKeeper.GetBalance(f.Context(), stakers[0].address, bondDenom)

			err = f.tallyKeeper.EndBlock(f.Context())
			require.NoError(t, err)
			require.NotContains(t, f.logBuf.String(), "ERR")

			// TODO query get_staker pending_withdrawal and check diff
			afterBalance := f.bankKeeper.GetBalance(f.Context(), stakers[0].address, bondDenom)
			diff := afterBalance.Sub(beforeBalance)
			require.Equal(t, "0aseda", diff.String())

			dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, tt.expExitCode, dataResult.ExitCode)

			dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
			require.NoError(t, err)
			require.Contains(t, dataResults, *dataResult)
		})
	}
}

func TestEndBlock_NoTallyReadyDataRequests(t *testing.T) {
	f := initFixture(t)
	err := f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)
	require.NotContains(t, f.logBuf.String(), "ERR")
}

func TestEndBlock_UpdateMaxResultSize(t *testing.T) {
	f := initFixture(t)

	// Set max result size to 1 and verify that the data request fails
	params := types.DefaultParams()
	params.MaxResultSize = 1
	msg := &types.MsgUpdateParams{
		Authority: f.tallyKeeper.GetAuthority(),
		Params:    params,
	}

	_, err := f.tallyMsgServer.UpdateParams(f.Context(), msg)
	require.NoError(t, err)

	drID, _ := f.commitRevealDataRequest(
		t, 1, 1, 1, false,
		commitRevealConfig{
			requestMemo: base64.StdEncoding.EncodeToString([]byte("memo")),
			reveal:      base64.StdEncoding.EncodeToString([]byte("reveal")),
		})

	err = f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
	require.NoError(t, err)
	require.NotEqual(t, uint32(0), dataResult.ExitCode)
	require.Contains(t, string(dataResult.Result), "Result larger than 1bytes")

	dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Contains(t, dataResults, *dataResult)

	// Set max result size to 1024 and verify that the data request succeeds
	params.MaxResultSize = 1024
	msg = &types.MsgUpdateParams{
		Authority: f.tallyKeeper.GetAuthority(),
		Params:    params,
	}

	_, err = f.tallyMsgServer.UpdateParams(f.Context(), msg)
	require.NoError(t, err)

	drID, _ = f.commitRevealDataRequest(
		t, 1, 1, 1, false,
		commitRevealConfig{
			requestMemo: base64.StdEncoding.EncodeToString([]byte("memo")),
			reveal:      base64.StdEncoding.EncodeToString([]byte("reveal")),
		})

	err = f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResultAfter, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
	require.NoError(t, err)
	require.Equal(t, uint32(0), dataResultAfter.ExitCode)
	require.Contains(t, string(dataResultAfter.Result), "VM_MODE=tally")

	dataResultsAfter, err := f.batchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Contains(t, dataResultsAfter, *dataResultAfter)
}

func TestEndBlock_PausedContract(t *testing.T) {
	f := initFixture(t)
	stakers := f.addStakers(t, 5)

	noCommitsDr, err := f.postDataRequest([]byte{}, []byte{}, base64.StdEncoding.EncodeToString([]byte("noCommits")), 1)
	require.NoError(t, err)

	noRevealsDr, err := f.postDataRequest([]byte{}, []byte{}, base64.StdEncoding.EncodeToString([]byte("noReveals")), 1)
	require.NoError(t, err)

	_, err = f.commitDataRequest(
		stakers, noRevealsDr.Height, noRevealsDr.DrID, 1,
		commitRevealConfig{
			reveal: base64.StdEncoding.EncodeToString([]byte("sike")),
		})
	require.NoError(t, err)

	resolvedDr, err := f.postDataRequest([]byte{}, []byte{}, base64.StdEncoding.EncodeToString([]byte("resolved")), 1)
	require.NoError(t, err)

	commitment, err := f.commitDataRequest(
		stakers, resolvedDr.Height, resolvedDr.DrID, 1,
		commitRevealConfig{
			reveal: base64.StdEncoding.EncodeToString([]byte("sike")),
		})
	require.NoError(t, err)

	err = f.revealDataRequest(
		stakers, resolvedDr.Height, resolvedDr.DrID, commitment, 1,
		commitRevealConfig{
			reveal: base64.StdEncoding.EncodeToString([]byte("sike")),
		})
	require.NoError(t, err)

	// Ensure the DR without commitments and the DR without reveals are timed out
	for i := 0; i < defaultRevealTimeoutBlocks; i++ {
		f.AddBlock()
	}

	f.pauseContract(t)

	err = f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)
	require.NotContains(t, f.logBuf.String(), "ERR")

	noCommitsResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), noCommitsDr.DrID)
	require.NoError(t, err)
	require.Equal(t, uint32(types.TallyExitCodeContractPaused), noCommitsResult.ExitCode)

	noRevealsResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), noRevealsDr.DrID)
	require.NoError(t, err)
	require.Equal(t, uint32(types.TallyExitCodeContractPaused), noRevealsResult.ExitCode)

	resolvedResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), resolvedDr.DrID)
	require.NoError(t, err)
	require.Equal(t, uint32(types.TallyExitCodeContractPaused), resolvedResult.ExitCode)
}
