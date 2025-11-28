package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestRevealWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateStakedTestAccount("bob", 22, 10)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 2)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits
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
	require.NoError(t, err)

	// bob commits
	bobReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	bobRevealMsg := bob.CreateRevealMsg(bobReveal)
	_, err = bob.CommitResult(bobRevealMsg)
	require.NoError(t, err)

	// alice reveals
	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)

	revealingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_REVEALING, 10, nil)
	require.NoError(t, err)
	require.Len(t, revealingDrsResp.DataRequests, 1)
	require.Len(t, revealingDrsResp.DataRequests[0].Reveals, 1)

	aliceBalance := alice.Balance()
	require.Equal(t, f.SedaToAseda(12), aliceBalance)
}

func TestWorksWithProxies(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	proxy1 := f.CreateProxyAccount("proxy1")

	// alice commits with a proxy
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{proxy1.PublicKeyHex()},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// alice reveals with a proxy
	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)

	tallyingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 10, nil)
	require.NoError(t, err)
	require.Len(t, tallyingDrsResp.DataRequests, 1)
	require.Len(t, tallyingDrsResp.DataRequests[0].Reveals, 1)
}

func TestFailsProxyWhenInvalidHex(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits with an invalid proxy
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{"invalidhex"},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// alice reveals with an invalid proxy
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "invalid hex-encoded proxy public key")
}

func TestNoErrorWhenNotRealProxyKey(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits with an invalid proxy
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{"deadbeef"},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// alice reveals with an invalid proxy
	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)
}

func TestFailsIfNotInRevealPhase(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice tries to reveal without committing first
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "reveal phase has not started")
}

func TestFailsIfRevealTimedOut(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// commit and move to reveal phase
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
	require.NoError(t, err)

	// advance past reveal timeout
	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)
	f.AdvanceBlocks(int64(drConfigResp.DataRequestConfig.RevealTimeoutInBlocks + 1))

	// alice tries to reveal
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "not found")

	// check data result
	dataResult, err := f.BatchingKeeper.GetDataResult(f.Context(), postDrResult.DrID, uint64(postDrResult.Height))
	require.NoError(t, err)
	require.Equal(t, types.TallyExitCodeFilterError, dataResult.ExitCode)
}

func TestRevealFailsOnExpiredDr(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// commit and reveal
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
	require.NoError(t, err)

	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)

	f.AdvanceBlocks(1)

	// alice tries to reveal again
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "not found")
}

func TestFailsIfUserDidNotCommit(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateStakedTestAccount("bob", 22, 10)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits moving dr to reveal phase
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
	require.NoError(t, err)

	// bob tries to reveal without committing first
	bobReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	bobRevealMsg := bob.CreateRevealMsg(bobReveal)
	_, err = bob.RevealResult(bobRevealMsg)
	require.ErrorContains(t, err, "commit under given public key does not exist")
}

func TestFailsOnDoubleReveal(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateStakedTestAccount("bob", 22, 10)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 2)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits
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
	require.NoError(t, err)

	// bob commits
	bobReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	bobRevealMsg := bob.CreateRevealMsg(bobReveal)
	_, err = bob.CommitResult(bobRevealMsg)
	require.NoError(t, err)

	// alice reveals
	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)

	// alice tries to reveal again
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "reveal under given public key already exists")
}

func TestFailsIfRevealDoesNotMatchCommit(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits
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
	require.NoError(t, err)

	// alice tries to reveal with a different reveal value
	aliceReveal.Reveal = testutil.RevealHelperFromString("20")
	aliceRevealMsg = alice.CreateRevealMsg(aliceReveal)
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "revealed result does not match the committed result")
}

func TestFailsIfProxyPubKeysChangeBetweenPhases(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	proxy1 := f.CreateProxyAccount("proxy1")
	proxy2 := f.CreateProxyAccount("proxy2")

	// alice commits with a proxy
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{proxy1.PublicKeyHex()},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// alice tries to reveal with a different proxy
	aliceReveal.ProxyPubKeys = []string{proxy2.PublicKeyHex()}
	aliceRevealMsg = alice.CreateRevealMsg(aliceReveal)
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "revealed result does not match the committed result")
}

func TestWorksAfterUnstaking(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits
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
	require.NoError(t, err)

	// alice unstakes
	_, err = alice.Unstake()
	require.NoError(t, err)

	// alice reveals
	_, err = alice.RevealResult(aliceRevealMsg)
	require.NoError(t, err)

	tallyingDrsResp, err := bob.GetDataRequestsByStatus(types.DATA_REQUEST_STATUS_TALLYING, 10, nil)
	require.NoError(t, err)
	require.Len(t, tallyingDrsResp.DataRequests, 1)
	require.Len(t, tallyingDrsResp.DataRequests[0].Reveals, 1)
}

func TestCannotFrontRunCommitReveal(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)
	charlie := f.CreateStakedTestAccount("charlie", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// charlie commits using alice's reveal
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = charlie.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// fred reveals using alice's reveal
	_, err = charlie.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "commit under given public key does not exist")
}

func TestCannontFrontRunReveal(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)
	charlie := f.CreateStakedTestAccount("charlie", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = charlie.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// charlie tries to reveal using alice's reveal
	charlieRevealMsg := charlie.CreateRevealMsg(aliceReveal)
	_, err = charlie.RevealResult(charlieRevealMsg)
	require.ErrorContains(t, err, "revealed result does not match the committed result")
}

func TestFailsWhenRevealTooBig(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateTestAccount("bob", 22)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)
	reveal := make([]byte, drConfigResp.DataRequestConfig.DrRevealSizeLimitInBytes+1)

	// alice commits
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        reveal,
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)

	// alice tries to reveal with a too big reveal
	aliceRevealMsg = alice.CreateRevealMsg(aliceReveal)
	_, err = alice.RevealResult(aliceRevealMsg)
	require.ErrorContains(t, err, "reveal is too big")
}

func TestFailsWhenRevealTooBigAccountedForRf(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	bob := f.CreateStakedTestAccount("bob", 22, 10)
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// post a dr
	dr := bob.CreatePostDRMsg("1", 2)
	postDrResult, err := bob.PostDataRequest(dr, nil)
	require.NoError(t, err)

	// alice commits
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
	require.NoError(t, err)

	drConfigResp, err := bob.GetDataRequestConfig()
	require.NoError(t, err)
	reveal := make([]byte, (drConfigResp.DataRequestConfig.DrRevealSizeLimitInBytes/2)+1)

	// bob commits
	bobReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        reveal,
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	bobRevealMsg := bob.CreateRevealMsg(bobReveal)
	_, err = bob.CommitResult(bobRevealMsg)
	require.NoError(t, err)

	// bob tries to reveal with a too big reveal
	_, err = bob.RevealResult(bobRevealMsg)
	require.ErrorContains(t, err, "reveal is too big")
}
