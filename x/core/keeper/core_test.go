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
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func TestEndBlock(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 5)

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
			err := f.SetDataProxyConfig(proxyPubKeys[0], "seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f", sdk.NewCoin(testutil.BondDenom, math.NewInt(1000000000000000000)))
			require.NoError(t, err)

			dr := testutil.NewTestDRWithRandomPrograms(
				[]byte("reveal"),   // reveal
				tt.memo,            // memo
				150000000000000000, // gas used
				0,                  // exit code
				proxyPubKeys,
				tt.replicationFactor,
				f.Context().BlockTime(),
			)

			drID := dr.ExecuteDataRequestFlow(f, tt.numCommits, tt.numReveals, tt.timeout)

			// Data request should be in the tallying status.
			drCheck, err := f.CoreKeeper.GetDataRequest(f.Context(), drID)
			require.NoError(t, err)

			if tt.timeout {
				if tt.numCommits >= tt.replicationFactor {
					require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, drCheck.Status)
				} else {
					require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, drCheck.Status)
				}
			} else {
				require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, drCheck.Status)
			}

			revealers, err := f.CoreKeeper.GetRevealers(f.Context(), drID)
			require.NoError(t, err)
			for _, revealer := range revealers {
				_, err := f.CoreKeeper.GetRevealBody(f.Context(), drID, revealer)
				require.NoError(t, err)
			}

			err = f.CoreKeeper.EndBlock(f.Context())
			require.NoError(t, err)

			// Data request should have been removed from the store.
			drCheck, err = f.CoreKeeper.GetDataRequest(f.Context(), drID)
			require.Error(t, err)

			revealersAfter, err := f.CoreKeeper.GetRevealers(f.Context(), drID)
			require.NoError(t, err)
			require.Empty(t, revealersAfter)

			for _, revealer := range revealers {
				_, err := f.CoreKeeper.GetRevealBody(f.Context(), drID, revealer)
				require.Error(t, err)
			}

			dataResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), drID)
			require.NoError(t, err)
			require.Equal(t, tt.expExitCode, dataResult.ExitCode)
		})
	}
}

func TestTxFeeRefund(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 5)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())

	tests := []struct {
		name               string
		CommitGasLimit     uint64
		RevealGasLimit     uint64
		commitExpectRefund bool
		revealExpectRefund bool
	}{
		{
			name:               "tx fees should be refunded",
			CommitGasLimit:     100000,
			RevealGasLimit:     100000,
			commitExpectRefund: true,
			revealExpectRefund: true,
		},
		{
			name:               "commit gas limit too large for refund",
			CommitGasLimit:     300000,
			RevealGasLimit:     100000,
			commitExpectRefund: false,
			revealExpectRefund: true,
		},
		{
			name:               "reveal gas limit too large for refund",
			CommitGasLimit:     150000,
			RevealGasLimit:     300000,
			commitExpectRefund: true,
			revealExpectRefund: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialBalance := f.BankKeeper.GetBalance(f.Context(), f.Stakers[0].Address, testutil.BondDenom)

			dr := testutil.NewTestDR(
				execProgram.Hash,
				tallyProgram.Hash,
				[]byte("reveal"), // reveal
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))), // memo
				150000000000000000, // gas used
				0,                  // exit code
				[]string{},         // proxy pub keys
				1,                  // replication factor
			)
			dr.SetCommitRevealGasLimits(tt.CommitGasLimit, tt.RevealGasLimit)

			dr.PostDataRequest(f)

			// Commit and check balance
			dr.CommitDataRequest(f, 1, nil)

			afterCommitBalance := f.BankKeeper.GetBalance(f.Context(), f.Stakers[0].Address, testutil.BondDenom)
			diff := initialBalance.Sub(afterCommitBalance)
			if tt.commitExpectRefund {
				require.Equal(t, "0aseda", diff.String(), "tx fee must have been refunded")
			} else {
				fee := sdk.NewCoins(sdk.NewCoin(testutil.BondDenom, math.NewIntFromUint64(tt.CommitGasLimit).Mul(math.NewInt(1e10))))
				require.Equal(t, fee.String(), diff.String(), "tx fee must have been deducted")
			}

			// Reveal and check balance
			dr.ExecuteReveals(f, 1, nil)

			afterRevealBalance := f.BankKeeper.GetBalance(f.Context(), f.Stakers[0].Address, testutil.BondDenom)
			diff = afterCommitBalance.Sub(afterRevealBalance)
			if tt.revealExpectRefund {
				require.Equal(t, "0aseda", diff.String(), "tx fee must have been refunded")
			} else {
				fee := sdk.NewCoins(sdk.NewCoin(testutil.BondDenom, math.NewIntFromUint64(tt.RevealGasLimit).Mul(math.NewInt(1e10))))
				require.Equal(t, fee.String(), diff.String(), "tx fee must have been deducted")
			}
		})
	}
}

func TestEndBlock_NoTallyReadyDataRequests(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	err := f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)
	require.NotContains(t, f.LogBuf.String(), "ERR")
}

func TestEndBlock_UpdateMaxResultSize(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 1)

	// Set max result size to 1 and verify that the data request fails
	params := types.DefaultParams()
	params.TallyConfig.MaxResultSize = 1
	owner, err := f.CoreKeeper.GetOwner(f.Context())
	require.NoError(t, err)
	msg := &types.MsgUpdateParams{
		Owner:  owner,
		Params: params,
	}

	_, err = f.CoreMsgServer.UpdateParams(f.Context(), msg)
	require.NoError(t, err)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm2(), f.Context().BlockTime())
	dr := testutil.NewTestDR(
		execProgram.Hash, tallyProgram.Hash,
		[]byte("reveal"),
		base64.StdEncoding.EncodeToString([]byte("memo")),
		150000000000000000, // gas used
		0,                  // exit code
		[]string{},         // proxy pub keys
		1,                  // replication factor
	)

	drID := dr.ExecuteDataRequestFlow(f, 1, 1, false)

	err = f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), drID)
	require.NoError(t, err)
	require.NotEqual(t, uint32(0), dataResult.ExitCode)
	require.Contains(t, string(dataResult.Result), "Result larger than 1bytes")

	dataResults, err := f.BatchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Contains(t, dataResults, *dataResult)

	// Ensure the new DR gets a unique ID
	f.AddBlock()

	// Set max result size to 1024 and verify that the data request succeeds
	params.TallyConfig.MaxResultSize = 1024
	msg = &types.MsgUpdateParams{
		Owner:  owner,
		Params: params,
	}

	_, err = f.CoreMsgServer.UpdateParams(f.Context(), msg)
	require.NoError(t, err)

	drID = dr.ExecuteDataRequestFlow(f, 1, 1, false)

	err = f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResultAfter, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), drID)
	require.NoError(t, err)
	require.Equal(t, uint32(0), dataResultAfter.ExitCode)
	require.Contains(t, string(dataResultAfter.Result), "VM_MODE=tally")

	dataResultsAfter, err := f.BatchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Contains(t, dataResultsAfter, *dataResultAfter)
}

// TestTallyTestItems executes the 100 randomly selected tally programs and
// verifies their results in the batching module store.
func TestTallyTestItems(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 1)

	numRequests := 100
	drIDs := make([]string, numRequests)
	testItems := make([]testwasms.TallyTestItem, numRequests)
	for i := range numRequests {
		randIndex := rand.Intn(len(testwasms.TestWasms))
		execProgram := wasmstoragetypes.NewOracleProgram(testwasms.TestWasms[randIndex], f.Context().BlockTime())

		randIndex = rand.Intn(len(testwasms.TallyTestItems))
		testItem := testwasms.TallyTestItems[randIndex]
		tallyProgram := wasmstoragetypes.NewOracleProgram(testItem.TallyProgram, f.Context().BlockTime())

		dr := testutil.NewTestDR(
			execProgram.Hash, tallyProgram.Hash,
			testItem.Reveal,
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", i))), // memo
			testItem.GasUsed,
			0,
			[]string{},
			1,
		)
		drIDs[i] = dr.ExecuteDataRequestFlow(f, 1, 1, false)
		testItems[i] = testItem
	}

	err := f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	// Check the batching module store for the data results.
	dataResults, err := f.BatchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Equal(t, numRequests, len(dataResults))

	for i, drID := range drIDs {
		dataResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), drID)
		require.NoError(t, err)

		require.Equal(t, 64, len(dataResult.Id))
		require.Equal(t, uint64(f.Context().BlockHeight()), dataResult.DrBlockHeight)
		require.Equal(t, uint64(f.Context().BlockHeight()), dataResult.BlockHeight)
		require.Equal(t, uint64(f.Context().BlockHeader().Time.Unix()), dataResult.BlockTimestamp)

		require.Equal(t, testItems[i].ExpectedResult, dataResult.Result)
		require.Equal(t, testItems[i].ExpectedExitCode, dataResult.ExitCode)
		// TODO Revive
		// require.Equal(t, testItems[i].ExpectedGasUsed.String(), dataResult.GasUsed.String())

		require.Contains(t, dataResults, *dataResult)
	}
}

func TestEndBlock_PausedCore(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 5)
	zeroHash := make([]byte, 32)

	// When the module is paused the DR poster should get a full refund and end up with the same balance as before posting
	beforeBalance := f.BankKeeper.GetBalance(f.Context(), f.Creator.AccAddress(), testutil.BondDenom)
	t.Log("Poster balance before posting:", beforeBalance, f.Creator.Address())

	noCommitsDR := testutil.NewTestDR(
		zeroHash, zeroHash,
		[]byte("sike"),
		base64.StdEncoding.EncodeToString([]byte("noCommits")),
		150000000000000000,
		0,
		[]string{},
		1,
	)
	noCommitsDR.PostDataRequest(f)

	noRevealsDR := testutil.NewTestDR(
		zeroHash, zeroHash,
		[]byte("sike"),
		base64.StdEncoding.EncodeToString([]byte("noReveals")),
		150000000000000000,
		0,
		[]string{},
		1,
	)
	noRevealsDR.PostDataRequest(f)
	noRevealsDR.CommitDataRequest(f, 1, nil)

	resolvedDR := testutil.NewTestDR(
		zeroHash, zeroHash,
		[]byte("sike"),
		base64.StdEncoding.EncodeToString([]byte("resolved")),
		150000000000000000,
		0,
		[]string{},
		1,
	)
	resolvedDR.PostDataRequest(f)
	resolvedDR.CommitDataRequest(f, 1, nil)
	resolvedDR.ExecuteReveals(f, 1, nil)

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

			noRevealsResult, err = f.BatchingKeeper.GetLatestDataResult(f.Context(), noRevealsDR.GetDataRequestID())
			require.NoError(t, err)
			require.Equal(t, uint32(types.TallyExitCodeContractPaused), noRevealsResult.ExitCode)

			resolvedResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), resolvedDR.GetDataRequestID())
			require.NoError(t, err)
			require.Equal(t, uint32(types.TallyExitCodeContractPaused), resolvedResult.ExitCode)
		}
	}

	err = f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)
	require.NotContains(t, f.LogBuf.String(), "ERR")

	noCommitsResult, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), noCommitsDR.GetDataRequestID())
	require.NoError(t, err)
	require.NotEqual(t, "", noCommitsResult.Id, "Result ID should not be empty")
	require.Equal(t, uint32(types.TallyExitCodeContractPaused), noCommitsResult.ExitCode)

	// Ensure the DR without reveals was removed from the module and not processed again
	noRevealsResultAfterTimeout, err := f.BatchingKeeper.GetLatestDataResult(f.Context(), noRevealsDR.GetDataRequestID())
	require.NoError(t, err)
	require.Equal(t, int(noRevealsResult.BlockHeight), int(noRevealsResultAfterTimeout.BlockHeight), "Already resolved DR was processed again")

	// Ensure the poster got a full refund for all the posted DRs.
	afterProcessingBalance := f.BankKeeper.GetBalance(f.Context(), f.Creator.AccAddress(), testutil.BondDenom)
	diff := afterProcessingBalance.Sub(beforeBalance)
	require.Equal(t, "0aseda", diff.String())
}
