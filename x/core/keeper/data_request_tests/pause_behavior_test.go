package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestPausePropertyDrQueryByStatus(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	_ = f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CalculateDrIdAndArgs("1", 1)
	_, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	drsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.False(t, drsResp.IsPaused)
	require.Len(t, drsResp.DataRequests, 1)

	_, err = f.Creator.Pause()
	require.NoError(t, err)

	drsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.True(t, drsResp.IsPaused)
	require.Len(t, drsResp.DataRequests, 1)

	_, err = f.Creator.Unpause()
	require.NoError(t, err)

	drsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.False(t, drsResp.IsPaused)
	require.Len(t, drsResp.DataRequests, 1)
}

func TestDataRequestTxsArePaused(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice is a Staker
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// pause the module
	_, err := f.Creator.Pause()
	require.NoError(t, err)

	// post a dr should fail
	dr := bob.CalculateDrIdAndArgs("1", 1)
	_, err = bob.PostDataRequest(dr, 1, nil)
	require.ErrorContains(t, err, "module is paused")

	// unpause the module
	_, err = f.Creator.Unpause()
	require.NoError(t, err)

	// post a dr
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// pause the module again
	_, err = f.Creator.Pause()
	require.NoError(t, err)

	// commit should fail
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.ErrorContains(t, err, "module is paused")

	// unpause the module again
	_, err = f.Creator.Unpause()
	require.NoError(t, err)

	// commit should work now
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// pause the module again
	_, err = f.Creator.Pause()
	require.NoError(t, err)

	// reveal should fail
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "module is paused")

	// unpause the module again
	_, err = f.Creator.Unpause()
	require.NoError(t, err)

	// reveal should work now
	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)

	// pause the module again
	_, err = f.Creator.Pause()
	require.NoError(t, err)

	// should still be able to update data request config
	newConfig := types.DefaultParams().DataRequestConfig
	newConfig.MemoLimitInBytes = 1
	_, err = f.Creator.SetDataRequestConfig(*newConfig)
	require.NoError(t, err)

	// fetch the config
	configResp, err := f.Creator.GetDataRequestConfig()
	require.NoError(t, err)
	require.Equal(t, uint32(1), configResp.DataRequestConfig.MemoLimitInBytes)
}
