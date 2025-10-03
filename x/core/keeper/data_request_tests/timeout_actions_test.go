package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

// TODO: Not sure how to test this... since advancing(calling endblock immediately would tally them?)
func TestTimedOutRequestsMoveToTally(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CalculateDrIdAndArgs("1", 1)
	_, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	drConfigResp, err := f.Creator.GetDataRequestConfig()
	require.NoError(t, err)

	// expire commit period
	f.AdvanceBlocks(int64(drConfigResp.DataRequestConfig.CommitTimeoutInBlocks))

	tallyDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 10, nil)
	require.NoError(t, err)
	require.Len(t, tallyDrsResp.DataRequests, 1)

	// post another dr
	dr2 := bob.CalculateDrIdAndArgs("2", 1)
	postDr2Result, err := bob.PostDataRequest(dr2, 1, nil)
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

	tallyDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 10, nil)
	require.NoError(t, err)
	require.Len(t, tallyDrsResp.DataRequests, 1)
}
