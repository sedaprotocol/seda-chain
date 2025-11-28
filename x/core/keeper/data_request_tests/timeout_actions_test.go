package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestTimedOutRequestsMoveToTally(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDr1Result, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	drConfigResp, err := f.Creator.GetDataRequestConfig()
	require.NoError(t, err)

	// expire commit period
	f.AdvanceBlocks(int64(drConfigResp.DataRequestConfig.CommitTimeoutInBlocks))

	// check that the dr is not in commit status
	commitDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, commitDrsResp.DataRequests, 0)

	// check that dr is not in reveal status
	revealDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 10, nil)
	require.NoError(t, err)
	require.Len(t, revealDrsResp.DataRequests, 0)

	// check that the dr is not in tally
	tallyDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 10, nil)
	require.NoError(t, err)
	require.Len(t, tallyDrsResp.DataRequests, 0)

	// check that dr has a result posted
	dataResult1, err := f.BatchingKeeper.GetDataResult(f.Context(), postDr1Result.DrID, uint64(postDr1Result.Height))
	require.NoError(t, err)
	require.NotNil(t, dataResult1)

	// post another dr
	dr2 := bob.CreatePostDRMsg("2", 1)
	postDr2Result, err := bob.PostDataRequest(dr2, nil)
	require.NoError(t, err)

	// Alice commits on second dr
	aliceReveal := &types.RevealBody{
		DrID:          postDr2Result.DrID,
		DrBlockHeight: uint64(postDr2Result.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// expire reveal period
	f.AdvanceBlocks(int64(drConfigResp.DataRequestConfig.RevealTimeoutInBlocks))

	// check that the dr is not in commit status
	commitDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, commitDrsResp.DataRequests, 0)

	// check that dr is not in reveal status
	revealDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 10, nil)
	require.NoError(t, err)
	require.Len(t, revealDrsResp.DataRequests, 0)

	// check that the dr is not in tally
	tallyDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 10, nil)
	require.NoError(t, err)
	require.Len(t, tallyDrsResp.DataRequests, 0)

	// check that dr has a result posted
	dataResult2, err := f.BatchingKeeper.GetDataResult(f.Context(), postDr2Result.DrID, uint64(postDr2Result.Height))
	require.NoError(t, err)
	require.NotNil(t, dataResult2)
}
