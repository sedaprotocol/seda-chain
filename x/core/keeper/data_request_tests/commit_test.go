package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestFailsIfNotStaked(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create a staker so a dr can be posted
	bob := f.CreateStakedTestAccount("bob", 22, 1)

	// Alice has no stake
	alice := f.CreateTestAccount("alice", 22)

	// Bob posts a data request
	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice tries to commit
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
	require.ErrorContains(t, err, "not found")
}

func TestFailsIfCommitTimedOut(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateStakedTestAccount("alice", 22, 1)
	bob := f.CreateTestAccount("bob", 22)

	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// move past the commit window
	drConfigResp, err := alice.GetDataRequestConfig()
	require.NoError(t, err)
	f.AdvanceBlocks(int64(drConfigResp.DataRequestConfig.CommitTimeoutInBlocks))

	// Alice tries to commit
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
	require.ErrorContains(t, err, "not found")

	// check data result
	dataResult, err := f.BatchingKeeper.GetDataResult(f.Context(), postDrResult.DrID, uint64(postDrResult.Height))
	require.NoError(t, err)
	require.Equal(t, types.TallyExitCodeNotEnoughCommits, dataResult.ExitCode)
}

func TestCommitFailsOnExpiredDr(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateStakedTestAccount("alice", 22, 1)
	bob := f.CreateTestAccount("bob", 22)

	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// move past the commit window
	drConfigResp, err := alice.GetDataRequestConfig()
	require.NoError(t, err)
	f.AdvanceBlocks(int64(drConfigResp.DataRequestConfig.CommitTimeoutInBlocks + 1))

	// Alice tries to commit
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
	require.ErrorContains(t, err, "not found")
}

func TestFailsIfNotEnoughStaked(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice has minimal stake
	alice := f.CreateStakedTestAccount("alice", 22, 1)

	config := types.DefaultParams().StakingConfig
	config.MinimumStake = f.SedaToAseda(10)
	f.Creator.SetStakingConfig(*config)

	// Bob posts a data request
	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice tries to commit
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
	require.ErrorContains(t, err, "stake amount is insufficient")
}

func TestCommitWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request
	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice tries to commit
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

	aliceBalance := alice.Balance()
	require.Equal(t, f.SedaToAseda(12), aliceBalance)
}

func TestMustMeetReplicationFactor(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateStakedTestAccount("bob", 22, 10)

	// Alice
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request
	dr := bob.CalculateDrIDAndArgs("1", 2)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice commits
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

	drResp, err := bob.GetDataRequest(postDrResult.DrID)
	require.NoError(t, err)
	require.Equal(t, postDrResult.DrID, drResp.DataRequest.ID)
	require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, drResp.DataRequest.Status)

	// Bob commits
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

	drResp, err = bob.GetDataRequest(postDrResult.DrID)
	require.NoError(t, err)
	require.Equal(t, postDrResult.DrID, drResp.DataRequest.ID)
	require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, drResp.DataRequest.Status)
}

func TestFailsDoubleCommit(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateStakedTestAccount("alice", 22, 1)
	bob := f.CreateStakedTestAccount("bob", 22, 10)

	dr := bob.CalculateDrIDAndArgs("1", 2)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice tries to commit
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

	// Alice tries to commit again
	_, err = alice.CommitResult(aliceRevealMsg)
	t.Log(err)
	require.ErrorContains(t, err, "commit under given public key already exists")
}

func TestFailsAfterRevealStarted(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create an account so a dr can be posted
	bob := f.CreateStakedTestAccount("bob", 22, 10)

	// Alice
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request
	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice tries to commit
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

	// Bob tries to commit after it has moved to reveal phase
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
	require.ErrorContains(t, err, "data request is not in committing state")
}

func TestWrongSignatureFails(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create a staker so a dr can be posted
	bob := f.CreateTestAccount("bob", 22)

	// Alice
	alice := f.CreateStakedTestAccount("alice", 22, 10)

	// Bob posts a data request
	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice tries to commit
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)

	msg := &types.MsgCommit{
		Sender:    alice.Address(),
		DrID:      aliceRevealMsg.RevealBody.DrID,
		Commit:    aliceRevealMsg.Proof,
		PublicKey: alice.PublicKeyHex(),
		Proof:     "deadbeef",
	}

	f.SetTx(100_000, alice.AccAddress(), msg)
	_, err = f.CoreMsgServer.Commit(f.Context(), msg)
	require.ErrorContains(t, err, "invalid commit proof")
}

func TestMustBeOnAllowlistToCommit(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateStakedTestAccount("alice", 22, 1)
	bob := f.CreateTestAccount("bob", 22)

	dr := bob.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Owner removes Alice from allowlist
	_, err = f.Creator.RemoveFromAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// Alice tries to commit
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
	require.ErrorContains(t, err, "public key is not in allowlist")
}
