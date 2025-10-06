package datarequesttests

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestEmptyWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)

	drsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, drsResp.DataRequests, 0)
}

func TestOneWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	_ = f.CreateStakedTestAccount("alice", 22, 1)

	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	drsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, drsResp.DataRequests, 1)
	require.Equal(t, postDrResult.DrID, drsResp.DataRequests[0].ID)
	require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, drsResp.DataRequests[0].Status)
}

func TestLimitWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	_ = f.CreateStakedTestAccount("alice", 22, 1)

	dr1 := bob.CalculateDrIDAndArgs("1", 1)
	dr2 := bob.CalculateDrIDAndArgs("2", 1)
	dr3 := bob.CalculateDrIDAndArgs("3", 1)

	_, err := bob.PostDataRequest(dr1, 1, nil)
	require.NoError(t, err)
	_, err = bob.PostDataRequest(dr2, 1, nil)
	require.NoError(t, err)
	_, err = bob.PostDataRequest(dr3, 1, nil)
	require.NoError(t, err)

	drsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 2, nil)
	require.NoError(t, err)
	require.Len(t, drsResp.DataRequests, 2)
}

func TestDrsAreSortedByGasPrice(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	_ = f.CreateStakedTestAccount("alice", 22, 1)

	dr1 := bob.CalculateDrIDAndArgs("1", 1)
	dr1.GasPrice = types.MinGasPrice.Add(math.NewInt(1))
	dr1Funds := (math.NewIntFromUint64(dr1.ExecGasLimit).Add(math.NewIntFromUint64(dr1.TallyGasLimit))).Mul(dr1.GasPrice)
	_, err := bob.PostDataRequest(dr1, 1, &dr1Funds)
	require.NoError(t, err)

	dr2 := bob.CalculateDrIDAndArgs("2", 1)
	dr2.GasPrice = types.MinGasPrice.Add(math.NewInt(3))
	dr2Funds := (math.NewIntFromUint64(dr2.ExecGasLimit).Add(math.NewIntFromUint64(dr2.TallyGasLimit))).Mul(dr2.GasPrice)
	_, err = bob.PostDataRequest(dr2, 1, &dr2Funds)
	require.NoError(t, err)

	dr3 := bob.CalculateDrIDAndArgs("3", 1)
	dr3.GasPrice = types.MinGasPrice.Add(math.NewInt(2))
	dr3Funds := (math.NewIntFromUint64(dr3.ExecGasLimit).Add(math.NewIntFromUint64(dr3.TallyGasLimit))).Mul(dr3.GasPrice)
	_, err = bob.PostDataRequest(dr3, 1, &dr3Funds)
	require.NoError(t, err)

	drsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, drsResp.DataRequests, 3)
	require.Equal(t, dr2.GasPrice, drsResp.DataRequests[0].GasPrice)
	require.Equal(t, dr3.GasPrice, drsResp.DataRequests[1].GasPrice)
	require.Equal(t, dr1.GasPrice, drsResp.DataRequests[2].GasPrice)
}

func TestDrsAreSortedByGasAndHeight(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)
	// start at block height 1
	f.AdvanceBlocks(1)

	bob := f.CreateTestAccount("bob", 22)
	_ = f.CreateStakedTestAccount("alice", 22, 1)

	dr1 := bob.CalculateDrIDAndArgs("1", 1)
	dr1.GasPrice = types.MinGasPrice.Add(math.NewInt(1))
	dr1Funds := (math.NewIntFromUint64(dr1.ExecGasLimit).Add(math.NewIntFromUint64(dr1.TallyGasLimit))).Mul(dr1.GasPrice)
	_, err := bob.PostDataRequest(dr1, 1, &dr1Funds)
	require.NoError(t, err)

	dr2 := bob.CalculateDrIDAndArgs("2", 1)
	dr2.GasPrice = types.MinGasPrice.Add(math.NewInt(10))
	dr2Funds := (math.NewIntFromUint64(dr2.ExecGasLimit).Add(math.NewIntFromUint64(dr2.TallyGasLimit))).Mul(dr2.GasPrice)
	_, err = bob.PostDataRequest(dr2, 1, &dr2Funds)
	require.NoError(t, err)

	f.AdvanceBlocks(1)

	dr3 := bob.CalculateDrIDAndArgs("3", 1)
	dr3.GasPrice = types.MinGasPrice.Add(math.NewInt(10))
	dr3Funds := (math.NewIntFromUint64(dr3.ExecGasLimit).Add(math.NewIntFromUint64(dr3.TallyGasLimit))).Mul(dr3.GasPrice)
	_, err = bob.PostDataRequest(dr3, 2, &dr3Funds)
	require.NoError(t, err)

	dr4 := bob.CalculateDrIDAndArgs("4", 1)
	dr4.GasPrice = types.MinGasPrice.Add(math.NewInt(2))
	dr4Funds := (math.NewIntFromUint64(dr4.ExecGasLimit).Add(math.NewIntFromUint64(dr4.TallyGasLimit))).Mul(dr4.GasPrice)
	_, err = bob.PostDataRequest(dr4, 2, &dr4Funds)
	require.NoError(t, err)

	drsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, drsResp.DataRequests, 4)
	require.Equal(t, dr2.GasPrice, drsResp.DataRequests[0].GasPrice)
	require.Equal(t, int64(1), drsResp.DataRequests[0].PostedHeight)
	require.Equal(t, dr3.GasPrice, drsResp.DataRequests[1].GasPrice)
	require.Equal(t, int64(2), drsResp.DataRequests[1].PostedHeight)
	require.Equal(t, dr4.GasPrice, drsResp.DataRequests[2].GasPrice)
	require.Equal(t, int64(2), drsResp.DataRequests[2].PostedHeight)
	require.Equal(t, dr1.GasPrice, drsResp.DataRequests[3].GasPrice)
	require.Equal(t, int64(1), drsResp.DataRequests[3].PostedHeight)
}

// TODO: Failing should not include the last seen index in the next page
func TestLastSeenIndexWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	_ = f.CreateStakedTestAccount("alice", 22, 1)

	dr1 := bob.CalculateDrIDAndArgs("1", 1)
	dr2 := bob.CalculateDrIDAndArgs("2", 1)
	dr3 := bob.CalculateDrIDAndArgs("3", 1)

	postDrResult1, err := bob.PostDataRequest(dr1, 1, nil)
	require.NoError(t, err)
	_, err = bob.PostDataRequest(dr2, 1, nil)
	require.NoError(t, err)
	_, err = bob.PostDataRequest(dr3, 1, nil)
	require.NoError(t, err)

	firstDrResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 1, nil)
	require.NoError(t, err)
	require.Len(t, firstDrResp.DataRequests, 1)
	require.Equal(t, postDrResult1.DrID, firstDrResp.DataRequests[0].ID)

	remainingDrResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, &firstDrResp.LastSeenIndex)
	require.NoError(t, err)
	require.Len(t, remainingDrResp.DataRequests, 2)

	var dr1Seen bool
	for _, dr := range remainingDrResp.DataRequests {
		if dr.ID == postDrResult1.DrID {
			dr1Seen = true
		}
	}
	require.False(t, dr1Seen)
}

func TestQueryByStatusManyDrsWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 2+25*20)
	alice := f.CreateStakedTestAccount("alice", 22, 1)

	for i := range 25 {
		dr := bob.CalculateDrIDAndArgs(fmt.Sprintf("%d", i), 1)
		postDrResult, err := bob.PostDataRequest(dr, 1, nil)
		require.NoError(t, err)

		aliceReveal := &types.RevealBody{
			DrID:          postDrResult.DrID,
			DrBlockHeight: uint64(postDrResult.Height),
			Reveal:        testutil.RevealHelperFromString("10"),
			GasUsed:       0,
			ExitCode:      0,
			ProxyPubKeys:  []string{},
		}
		aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)

		if i < 15 {
			_, err = alice.CommitResult(aliceRevealMsg)
			require.NoError(t, err)
		}

		if i < 3 {
			_, err = alice.RevealResult(aliceRevealMsg)
			require.NoError(t, err)
		}
	}

	committingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 10, nil)
	require.NoError(t, err)
	require.Len(t, committingDrsResp.DataRequests, 10)

	revealingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 15, nil)
	require.NoError(t, err)
	require.Len(t, revealingDrsResp.DataRequests, 12)

	tallyingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 15, nil)
	require.NoError(t, err)
	require.Len(t, tallyingDrsResp.DataRequests, 3)
}

func TestQueryByStatusManyMoreDrsWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 2+163*20)
	alice := f.CreateStakedTestAccount("alice", 22, 1)

	currentHeight := uint64(f.Context().BlockHeight())
	reveal := testutil.RevealHelperFromString("10")

	// post 100 drs, commit half of them, and post another 50 drs
	for i := range 100 {
		dr := bob.CalculateDrIDAndArgs(fmt.Sprintf("%d", i), 1)
		postDrResult, err := bob.PostDataRequest(dr, 1, nil)
		require.NoError(t, err)

		if i == 0 {
			currentHeight = uint64(postDrResult.Height)
		}

		aliceReveal := &types.RevealBody{
			DrID:          postDrResult.DrID,
			DrBlockHeight: currentHeight,
			Reveal:        reveal,
			GasUsed:       0,
			ExitCode:      0,
			ProxyPubKeys:  []string{},
		}
		aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)

		if i%2 == 0 {
			_, err = alice.CommitResult(aliceRevealMsg)
			require.NoError(t, err)

			anotherDR := bob.CalculateDrIDAndArgs(fmt.Sprintf("%d", i+20_000), 1)
			_, err = bob.PostDataRequest(anotherDR, 1, nil)
			require.NoError(t, err)
		}
	}

	committingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, committingDrsResp.DataRequests, 100)

	revealingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, revealingDrsResp.DataRequests, 50)

	tallyingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, tallyingDrsResp.DataRequests, 0)

	// loop over revealing drs and reveal a quarter of them
	// and also post more drs
	for i, dr := range revealingDrsResp.DataRequests {
		if i%4 == 0 {

			aliceReveal := &types.RevealBody{
				DrID:          dr.ID,
				DrBlockHeight: currentHeight,
				Reveal:        reveal,
				GasUsed:       0,
				ExitCode:      0,
				ProxyPubKeys:  []string{},
			}
			aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
			_, err = alice.RevealResult(aliceRevealMsg)
			require.NoError(t, err)

			anotherDR := bob.CalculateDrIDAndArgs(fmt.Sprintf("%d", i+10_000), 1)
			_, err = bob.PostDataRequest(anotherDR, 1, nil)
			require.NoError(t, err)

		}
	}

	committingDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, committingDrsResp.DataRequests, 113)

	revealingDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, revealingDrsResp.DataRequests, 37)

	tallyingDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, tallyingDrsResp.DataRequests, 13)

	// advance one block to move tallying drs to a resolved state
	f.AdvanceBlocks(1)

	committingDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_COMMITTING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, committingDrsResp.DataRequests, 113)

	revealingDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, revealingDrsResp.DataRequests, 37)

	tallyingDrsResp, err = bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 1000, nil)
	require.NoError(t, err)
	require.Len(t, tallyingDrsResp.DataRequests, 0)
}

func TestQueryStatusesWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateStakedTestAccount("bob", 22, 1)
	alice := f.CreateStakedTestAccount("alice", 22, 1)

	dr1 := bob.CalculateDrIDAndArgs("1", 2)
	dr2 := bob.CalculateDrIDAndArgs("2", 1)
	dr3 := bob.CalculateDrIDAndArgs("3", 1)

	postDrResult1, err := bob.PostDataRequest(dr1, 1, nil)
	require.NoError(t, err)
	postDrResult2, err := bob.PostDataRequest(dr2, 1, nil)
	require.NoError(t, err)
	postDrResult3, err := bob.PostDataRequest(dr3, 1, nil)
	require.NoError(t, err)

	drIds := []string{postDrResult1.DrID, postDrResult2.DrID, postDrResult3.DrID}

	drsResp, err := bob.GetDataRequestStatuses(drIds)
	require.NoError(t, err)
	require.Len(t, drsResp.Statuses, 3)
	for _, status := range drsResp.Statuses {
		require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, status.Value)
	}

	// upgrade one dr to revealing
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult1.DrID,
		DrBlockHeight: uint64(postDrResult1.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	bobReveal := &types.RevealBody{
		DrID:          postDrResult1.DrID,
		DrBlockHeight: uint64(postDrResult1.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	bobRevealMsg := bob.CreateRevealMsg(bobReveal)
	_, err = bob.CommitResult(bobRevealMsg)
	require.NoError(t, err)

	// upgrade the second dr to tally phase
	aliceReveal2 := &types.RevealBody{
		DrID:          postDrResult2.DrID,
		DrBlockHeight: uint64(postDrResult2.Height),
		Reveal:        testutil.RevealHelperFromString("20"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg2 := alice.CreateRevealMsg(aliceReveal2)
	_, err = alice.CommitResult(aliceRevealMsg2)
	require.NoError(t, err)
	_, err = alice.RevealResult(aliceRevealMsg2)
	require.NoError(t, err)

	// confirm the statuses
	statusesResp, err := bob.GetDataRequestStatuses(drIds)
	require.NoError(t, err)
	require.Len(t, statusesResp.Statuses, 3)
	require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, statusesResp.Statuses[postDrResult1.DrID].Value)
	require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, statusesResp.Statuses[postDrResult2.DrID].Value)
	require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, statusesResp.Statuses[postDrResult3.DrID].Value)

	// resolve the tally dr
	f.AdvanceBlocks(1)

	// once more, confirm the statuses
	statusesResp, err = bob.GetDataRequestStatuses(drIds)
	require.NoError(t, err)
	require.Len(t, statusesResp.Statuses, 3)
	require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, statusesResp.Statuses[postDrResult1.DrID].Value)
	require.Nil(t, statusesResp.Statuses[postDrResult2.DrID])
	require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, statusesResp.Statuses[postDrResult3.DrID].Value)
}

func TestQueryStatusesNonExistingDr(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateStakedTestAccount("bob", 22, 1)

	drIds := []string{"nonexistingdr"}

	resp, err := bob.GetDataRequestStatuses(drIds)
	require.NoError(t, err)
	require.Len(t, resp.Statuses, 1)
	require.Nil(t, resp.Statuses["nonexistingdr"])
}
