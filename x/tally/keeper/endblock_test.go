package keeper_test

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	tallykeeper "github.com/sedaprotocol/seda-chain/x/tally/keeper"
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

			drID := f.executeDataRequestFlow(
				t, nil, nil,
				tt.replicationFactor, tt.numCommits, tt.numReveals, tt.timeout,
				commitRevealConfig{
					requestHeight: 1,
					requestMemo:   tt.memo,
					reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
					proxyPubKeys:  proxyPubKeys,
					gasUsed:       150000000000000000,
				})

			beforeBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			posterBeforeBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)

			err = f.tallyKeeper.EndBlock(f.Context())
			require.NoError(t, err)

			// TODO query get_staker pending_withdrawal and check diff
			// Verify the staker did not pay for the transactions
			afterBalance := f.bankKeeper.GetBalance(f.Context(), f.stakers[0].address, bondDenom)
			diff := afterBalance.Sub(beforeBalance)
			require.Equal(t, "0aseda", diff.String())

			// Verify the poster paid for execution
			afterPostBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)
			diff = afterPostBalance.Sub(posterBeforeBalance)
			require.NotEqual(t, "0aseda", diff.String(), "Poster should have paid for execution")

			dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
			require.NoError(t, err)
			// TODO map oracle program to exit code
			// require.Equal(t, tt.expExitCode, dataResult.ExitCode)

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

	drID := f.executeDataRequestFlow(
		t, testwasms.SampleTallyWasm(), testwasms.SampleTallyWasm2(),
		1, 1, 1, false,
		commitRevealConfig{
			requestHeight: 1,
			requestMemo:   base64.StdEncoding.EncodeToString([]byte("memo")),
			reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
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

	// Ensure the new DR gets a unique ID
	f.AddBlock()

	// Set max result size to 1024 and verify that the data request succeeds
	params.MaxResultSize = 1024
	msg = &types.MsgUpdateParams{
		Authority: f.tallyKeeper.GetAuthority(),
		Params:    params,
	}

	_, err = f.tallyMsgServer.UpdateParams(f.Context(), msg)
	require.NoError(t, err)

	drID = f.executeDataRequestFlow(
		t, testwasms.SampleTallyWasm(), testwasms.SampleTallyWasm2(),
		1, 1, 1, false,
		commitRevealConfig{
			requestHeight: 1,
			requestMemo:   base64.StdEncoding.EncodeToString([]byte("memo")),
			reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
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

func TestEndBlock_ChunkedContractQuery(t *testing.T) {
	f := initFixture(t)

	// Create more data requests than the max number of data requests per query.
	numDataRequests := tallykeeper.MaxDataRequestsPerQuery + 5

	for range numDataRequests {
		f.executeDataRequestFlow(
			t, nil, nil,
			1, 1, 1, false,
			commitRevealConfig{
				requestHeight: 1,
				requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("memo-%d", rand.Uint64()))),
				reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
			})
	}

	err := f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Len(t, dataResults, int(numDataRequests))
}

func TestEndBlock_ChunkedContractQuery_MaxTalliesPerBlock(t *testing.T) {
	f := initFixture(t)

	// Create more data requests than the max number of data requests per query.
	numDataRequests := tallykeeper.MaxDataRequestsPerQuery + 10

	params := types.DefaultParams()
	params.MaxTalliesPerBlock = numDataRequests
	f.tallyKeeper.SetParams(f.Context(), params)

	for range numDataRequests {
		f.executeDataRequestFlow(
			t, nil, nil,
			1, 1, 1, false,
			commitRevealConfig{
				requestHeight: 1,
				requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("memo-%d", rand.Uint64()))),
				reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
			})
	}

	err := f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Len(t, dataResults, int(numDataRequests))
}

func TestEndBlock_ChunkedContractQuery_LowMaxTalliesPerBlock(t *testing.T) {
	f := initFixture(t)

	// Create less data requests than the max number of data requests per query.
	numDataRequests := tallykeeper.MaxDataRequestsPerQuery - 25

	params := types.DefaultParams()
	params.MaxTalliesPerBlock = 1
	f.tallyKeeper.SetParams(f.Context(), params)

	for range numDataRequests {
		f.executeDataRequestFlow(
			t, nil, nil,
			1, 1, 1, false,
			commitRevealConfig{
				requestHeight: 1,
				requestMemo:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("memo-%d", rand.Uint64()))),
				reveal:        base64.StdEncoding.EncodeToString([]byte("reveal")),
			})
	}

	err := f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Len(t, dataResults, 1)
}

func TestEndBlock_PausedContract(t *testing.T) {
	f := initFixture(t)
	stakers := f.addStakers(t, 5)
	zeroHash := make([]byte, 32)

	// When the contract is paused the DR poster should get a full refund and end up with the same balance as before posting
	beforeBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)

	noCommitsDr, err := f.postDataRequest(zeroHash, zeroHash, base64.StdEncoding.EncodeToString([]byte("noCommits")), 1)
	require.NoError(t, err)

	noRevealsDr, err := f.postDataRequest(zeroHash, zeroHash, base64.StdEncoding.EncodeToString([]byte("noReveals")), 1)
	require.NoError(t, err)

	_, err = f.commitDataRequest(
		stakers[:1], noRevealsDr.Height, noRevealsDr.DrID,
		commitRevealConfig{
			requestHeight: 1,
			reveal:        base64.StdEncoding.EncodeToString([]byte("sike")),
		})
	require.NoError(t, err)

	resolvedDr, err := f.postDataRequest(zeroHash, zeroHash, base64.StdEncoding.EncodeToString([]byte("resolved")), 1)
	require.NoError(t, err)

	revealMsgs, err := f.commitDataRequest(
		stakers[:1], resolvedDr.Height, resolvedDr.DrID,
		commitRevealConfig{
			requestHeight: 1,
			reveal:        base64.StdEncoding.EncodeToString([]byte("sike")),
		})
	require.NoError(t, err)

	err = f.executeReveals(stakers, revealMsgs)
	require.NoError(t, err)

	afterPostBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)
	require.True(t, afterPostBalance.IsLT(beforeBalance), "Poster should have escrowed funds")

	f.pauseContract(t)

	var noRevealsResult *batchingtypes.DataResult

	// Ensure the DR without commitments and the DR without reveals are timed out
	for i := range defaultCommitTimeoutBlocks {
		f.AddBlock()

		// DRs in the reveal stage time out before DRs in the commit stage
		if i == defaultRevealTimeoutBlocks-1 {
			err = f.tallyKeeper.EndBlock(f.Context())
			require.NoError(t, err)
			require.NotContains(t, f.logBuf.String(), "ERR")

			noRevealsResult, err = f.batchingKeeper.GetLatestDataResult(f.Context(), noRevealsDr.DrID)
			require.NoError(t, err)
			require.Equal(t, uint32(types.TallyExitCodeContractPaused), noRevealsResult.ExitCode)

			resolvedResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), resolvedDr.DrID)
			require.NoError(t, err)
			require.Equal(t, uint32(types.TallyExitCodeContractPaused), resolvedResult.ExitCode)
		}
	}

	err = f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)
	require.NotContains(t, f.logBuf.String(), "ERR")

	noCommitsResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), noCommitsDr.DrID)
	require.NoError(t, err)
	require.NotEqual(t, "", noCommitsResult.Id, "Result ID should not be empty")
	require.Equal(t, uint32(types.TallyExitCodeContractPaused), noCommitsResult.ExitCode)

	// Ensure the DR without reveals was removed from the contract and not processed again
	noRevealsResultAfterTimeout, err := f.batchingKeeper.GetLatestDataResult(f.Context(), noRevealsDr.DrID)
	require.NoError(t, err)
	require.Equal(t, int(noRevealsResultAfterTimeout.BlockHeight), int(noRevealsResult.BlockHeight), "Already resolved DR was processed again")

	// Ensure the poster got a full refund for all the posted DRs.
	afterProcessingBalance := f.bankKeeper.GetBalance(f.Context(), f.deployer, bondDenom)
	diff := afterProcessingBalance.Sub(beforeBalance)
	require.Equal(t, "0aseda", diff.String())
}

// TestTallyTestItems executes the 100 randomly selected tally programs and
// verifies their results in the batching module store.
func TestTallyTestItems(t *testing.T) {
	f := initFixture(t)

	numRequests := 100
	drIDs := make([]string, numRequests)
	testItems := make([]testwasms.TallyTestItem, numRequests)
	for i := range numRequests {
		drIDs[i], testItems[i] = f.executeDataRequestFlowWithTallyTestItem(t, []byte(fmt.Sprintf("%d", i)))
	}

	err := f.tallyKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	// Check the batching module store for the data results.
	dataResults, err := f.batchingKeeper.GetDataResults(f.Context(), false)
	require.NoError(t, err)
	require.Equal(t, numRequests, len(dataResults))

	for i, drID := range drIDs {
		dataResult, err := f.batchingKeeper.GetLatestDataResult(f.Context(), drID)
		require.NoError(t, err)

		require.Equal(t, 64, len(dataResult.Id))
		require.Equal(t, uint64(f.Context().BlockHeight()), dataResult.DrBlockHeight)
		require.Equal(t, uint64(f.Context().BlockHeight()), dataResult.BlockHeight)
		require.Equal(t, uint64(f.Context().BlockHeader().Time.Unix()), dataResult.BlockTimestamp)

		require.Equal(t, testItems[i].ExpectedResult, dataResult.Result)
		require.Equal(t, testItems[i].ExpectedExitCode, dataResult.ExitCode)
		require.Equal(t, testItems[i].ExpectedGasUsed.String(), dataResult.GasUsed.String())

		require.Contains(t, dataResults, *dataResult)
	}
}
