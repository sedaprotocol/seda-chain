package keeper_test

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/stretchr/testify/require"
)

func TestStake(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "")
	require.NoError(t, err)

	// TODO: need to check the stake is recorded correctly
}

func TestStakeWithMemo(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "test memo")
	require.NoError(t, err)

	// TODO: need to check the stake is recorded correctly
}

func TestUnstake(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "")
	require.NoError(t, err)

	_, err = alice.Unstake()
	require.NoError(t, err)

	// TODO: need to check the unstake is recorded correctly
}

func TestWithdraw(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "")
	require.NoError(t, err)

	// TODO: need to give them rewards

	_, err = alice.Withdraw(nil)
	require.NoError(t, err)

	// TODO: need to check the withdraw is recorded correctly
}

func TestWithdrawRemovesStakerIfNoStake(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "")
	require.NoError(t, err)

	_, err = alice.Unstake()
	require.NoError(t, err)

	_, err = alice.Withdraw(nil)
	require.NoError(t, err)

	// TODO: need queries to check the staker is removed
}

func TestCannotStakeIfPaused(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	f.Creator.Pause()
	_, err := alice.Stake(10, "")
	require.Error(t, err)
}

func TestCannotUnstakeIfPaused(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "")
	require.NoError(t, err)
	f.Creator.Pause()

	_, err = alice.Unstake()
	require.Error(t, err)
}

func TestCannotWithdrawIfPaused(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10, "")
	require.NoError(t, err)
	f.Creator.Pause()
	_, err = alice.Withdraw(nil)
	require.Error(t, err)
}
