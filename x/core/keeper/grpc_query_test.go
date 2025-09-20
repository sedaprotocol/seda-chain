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
	f := testutil.InitFixture(t)
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

			config := testutil.CommitRevealConfig{
				RequestHeight: 1,
				RequestMemo:   "",
				Reveal:        []byte("reveal"),
				ProxyPubKeys:  []string{},
				GasUsed:       150000000000000000,
			}

			// Post and check
			postResults := make([]testutil.PostDataRequestResponse, tt.numPosts)
			for i := uint64(0); i < tt.numPosts; i++ {
				postResults[i] = f.PostDataRequest(
					execProgram.Hash, tallyProgram.Hash,
					base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
					1,
				)
			}
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_COMMITTING, tt.numPosts, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_REVEALING, 0, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_TALLYING, 0, fetchLimit)

			// Commit and check
			revealMsgs := make([][][]byte, tt.numCommits)
			for i, postRes := range postResults[:tt.numCommits] {
				revealMsgs[i] = f.CommitDataRequest(f.Stakers[:1], postRes.Height, postRes.DrID, config)
			}
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_COMMITTING, tt.numPosts-tt.numCommits, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_REVEALING, tt.numCommits, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_TALLYING, 0, fetchLimit)

			// Reveal and check
			for _, revealMsg := range revealMsgs[:tt.numReveals] {
				f.ExecuteReveals(f.Stakers[:1], revealMsg, config)
			}
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_COMMITTING, tt.numPosts-tt.numCommits, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_REVEALING, tt.numCommits-tt.numReveals, fetchLimit)
			f.CheckDataRequestsByStatus(t, types.DATA_REQUEST_STATUS_TALLYING, tt.numReveals, fetchLimit)
		})
	}
}

func TestGetExecutors(t *testing.T) {
	f := testutil.InitFixture(t)

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
