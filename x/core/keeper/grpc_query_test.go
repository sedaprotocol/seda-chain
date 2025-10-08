package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func TestGetDataRequestsByStatus(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 5)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())
	fetchLimit := uint64(100)

	tests := []struct {
		name       string
		numPosts   uint64
		numCommits uint64
		numReveals uint64
	}{
		{
			name:       "",
			numPosts:   862,
			numCommits: 481,
			numReveals: 250,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, tt.numPosts >= tt.numCommits)
			require.True(t, tt.numCommits >= tt.numReveals)

			// Post and check
			testDRs := make([]testutil.TestDR, tt.numPosts)
			for i := uint64(0); i < tt.numPosts; i++ {
				testDRs[i] = testutil.NewTestDR(
					execProgram.Hash, tallyProgram.Hash,
					[]byte("reveal"),
					base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
					150000000000000000,
					0,
					[]string{},
					1,
				)
				testDRs[i].PostDataRequest(f)
			}
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_COMMITTING, tt.numPosts, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_REVEALING, 0, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_TALLYING, 0, fetchLimit)

			// Commit and check
			for i := range testDRs[:tt.numCommits] {
				testDRs[i].CommitDataRequest(f, 1, nil)
			}
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_COMMITTING, tt.numPosts-tt.numCommits, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_REVEALING, tt.numCommits, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_TALLYING, 0, fetchLimit)

			// Reveal and check
			for _, testDR := range testDRs[:tt.numReveals] {
				testDR.ExecuteReveals(f, 1, nil)
			}
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_COMMITTING, tt.numPosts-tt.numCommits, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_REVEALING, tt.numCommits-tt.numReveals, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_TALLYING, tt.numReveals, fetchLimit)
		})
	}
}

func TestGetCommittersAndRevealers(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 15)

	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())

	committers, revealers, err := f.CoreKeeper.GetCommittersAndRevealers(f.Context(), "abcdef")
	require.NoError(t, err)
	require.NotNil(t, committers)
	require.NotNil(t, revealers)
	require.Equal(t, 0, len(committers))
	require.Equal(t, 0, len(revealers))

	// Post a DR with RF = 10.
	testDR := testutil.NewTestDR(
		execProgram.Hash, tallyProgram.Hash,
		[]byte("reveal"),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
		150000000000000000,
		0,
		[]string{},
		10,
	)
	testDR.PostDataRequest(f)

	committers, revealers, err = f.CoreKeeper.GetCommittersAndRevealers(f.Context(), testDR.GetDataRequestID())
	require.NoError(t, err)
	require.NotNil(t, committers)
	require.NotNil(t, revealers)
	require.Equal(t, 0, len(committers))
	require.Equal(t, 0, len(revealers))

	// Commit and check.
	testDR.CommitDataRequest(f, 7, nil)

	committers, revealers, err = f.CoreKeeper.GetCommittersAndRevealers(f.Context(), testDR.GetDataRequestID())
	require.NoError(t, err)
	require.Equal(t, 7, len(committers))
	require.Equal(t, 0, len(revealers))

	// The rest commit to meet the replication factor.
	testDR.CommitDataRequest(f, 3, []int{7, 8, 9})

	// Reveal and check.
	testDR.ExecuteReveals(f, 4, nil)

	committers, revealers, err = f.CoreKeeper.GetCommittersAndRevealers(f.Context(), testDR.GetDataRequestID())
	require.NoError(t, err)
	require.Equal(t, 10, len(committers))
	require.Equal(t, 4, len(revealers))
}

func TestGetExecutors(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	f.AddStakers(t, 5) // total 5 stakers

	executors, err := f.CoreKeeper.GetExecutors(f.Context(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 5, len(executors))
	for i := range executors {
		require.Equal(t, f.Stakers[i].PubKey, executors[i].GetPublicKey())
	}

	f.AddStakers(t, 23) // total 28 stakers

	executors, err = f.CoreKeeper.GetExecutors(f.Context(), 0, 10)
	require.NoError(t, err)
	require.Equal(t, 10, len(executors)) // query limit is 10
	i := 0
	for _, executor := range executors {
		require.Equal(t, f.Stakers[i].PubKey, executor.GetPublicKey())
		i++
	}

	executors, err = f.CoreKeeper.GetExecutors(f.Context(), 10, 10)
	require.NoError(t, err)
	require.Equal(t, 10, len(executors)) // query limit is 10
	for _, executor := range executors {
		require.Equal(t, f.Stakers[i].PubKey, executor.GetPublicKey())
		i++
	}

	executors, err = f.CoreKeeper.GetExecutors(f.Context(), 20, 10)
	require.NoError(t, err)
	require.Equal(t, 8, len(executors))
	for _, executor := range executors {
		require.Equal(t, f.Stakers[i].PubKey, executor.GetPublicKey())
		i++
	}

	// arbitrary offset and limit
	executors, err = f.CoreKeeper.GetExecutors(f.Context(), 17, 6)
	require.NoError(t, err)
	require.Equal(t, 6, len(executors))
	i = 17 // offset is 17
	for _, executor := range executors {
		require.Equal(t, f.Stakers[i].PubKey, executor.GetPublicKey())
		i++
	}
}
