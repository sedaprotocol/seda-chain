package keeper_test

import (
	"encoding/base64"
	"testing"

	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestEndBlock_PausedCore(t *testing.T) {
	f := testutil.InitFixture(t)
	stakers := f.AddStakers(t, 5)
	zeroHash := make([]byte, 32)

	// When the module is paused the DR poster should get a full refund and end up with the same balance as before posting
	beforeBalance := f.BankKeeper.GetBalance(f.Context(), f.Creator.AccAddress(), testutil.BondDenom)
	t.Log("Poster balance before posting:", beforeBalance, f.Creator.Address())

	noCommitsDr := f.PostDataRequest(zeroHash, zeroHash, base64.StdEncoding.EncodeToString([]byte("noCommits")), 1)

	noRevealsDr := f.PostDataRequest(zeroHash, zeroHash, base64.StdEncoding.EncodeToString([]byte("noReveals")), 1)

	config := testutil.CommitRevealConfig{
		RequestHeight: 1,
		Reveal:        []byte("sike"),
	}
	_ = f.CommitDataRequest(
		stakers[:1], noRevealsDr.Height, noRevealsDr.DrID,
		config,
	)
	resolvedDr := f.PostDataRequest(zeroHash, zeroHash, base64.StdEncoding.EncodeToString([]byte("resolved")), 1)

	revealMsgs := f.CommitDataRequest(
		stakers[:1], resolvedDr.Height, resolvedDr.DrID,
		config,
	)

	f.ExecuteReveals(stakers[:1], revealMsgs, config)

	afterPostBalance := f.BankKeeper.GetBalance(f.Context(), f.Creator.AccAddress(), testutil.BondDenom)
	require.True(t, afterPostBalance.IsLT(beforeBalance), "Poster should have escrowed funds")

	f.Creator.Pause()
	params, err := f.CoreKeeper.GetParams(f.Context())
	require.NoError(t, err)

	defaultCommitTimeoutBlocks := params.GetDataRequestConfig().CommitTimeoutInBlocks
	defaultRevealTimeoutBlocks := params.GetDataRequestConfig().RevealTimeoutInBlocks
	var noRevealsResult *batchingtypes.DataResult

	// Ensure the DR without commitments and the DR without reveals are timed out
	for i := range defaultCommitTimeoutBlocks {
		f.AddBlock()

		// DRs in the reveal stage time out before DRs in the commit stage
		if i == defaultRevealTimeoutBlocks-1 {
			err = f.CoreKeeper.EndBlock(f.Context())
			require.NoError(t, err)
			require.NotContains(t, f.LogBuf.String(), "ERR")

			noRevealsResult, err = f.BatchingKeeper.GetLatestDataResult(f.Context(), noRevealsDr.DrID)
			require.NoError(t, err)
			require.Equal(t, uint32(types.TallyExitCodeContractPaused), noRevealsResult.ExitCode)

			resolvedResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), resolvedDr.DrID)
			require.NoError(t, err)
			require.Equal(t, uint32(types.TallyExitCodeContractPaused), resolvedResult.ExitCode)
		}
	}

	err = f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)
	require.NotContains(t, f.LogBuf.String(), "ERR")

	noCommitsResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), noCommitsDr.DrID)
	require.NoError(t, err)
	require.NotEqual(t, "", noCommitsResult.Id, "Result ID should not be empty")
	require.Equal(t, uint32(types.TallyExitCodeContractPaused), noCommitsResult.ExitCode)

	// Ensure the DR without reveals was removed from the module and not processed again
	noRevealsResultAfterTimeout, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), noRevealsDr.DrID)
	require.NoError(t, err)
	require.Equal(t, int(noRevealsResult.BlockHeight), int(noRevealsResultAfterTimeout.BlockHeight), "Already resolved DR was processed again")

	// Ensure the poster got a full refund for all the posted DRs.
	afterProcessingBalance := f.BankKeeper.GetBalance(f.Context(), f.Creator.AccAddress(), testutil.BondDenom)
	diff := afterProcessingBalance.Sub(beforeBalance)
	require.Equal(t, "0aseda", diff.String())
}
