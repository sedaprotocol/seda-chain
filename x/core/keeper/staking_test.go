package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/stretchr/testify/require"
)

func TestStake(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	aliceStaker, err := alice.GetStaker()
	require.NoError(t, err)
	require.Equal(t, alice.PublicKeyHex(), aliceStaker.Staker.PublicKey)
	require.Equal(t, f.SedaToAseda(10), aliceStaker.Staker.Staked)
	require.Equal(t, f.SedaToAseda(0), aliceStaker.Staker.PendingWithdrawal)
	require.Equal(t, uint64(1), aliceStaker.Staker.SequenceNum)
	require.Equal(t, "", aliceStaker.Staker.Memo)

	// can stake again
	_, err = alice.Stake(5)
	require.NoError(t, err)

	aliceStaker, err = alice.GetStaker()
	require.NoError(t, err)
	require.Equal(t, alice.PublicKeyHex(), aliceStaker.Staker.PublicKey)
	require.Equal(t, f.SedaToAseda(15), aliceStaker.Staker.Staked)
	require.Equal(t, f.SedaToAseda(0), aliceStaker.Staker.PendingWithdrawal)
	require.Equal(t, uint64(2), aliceStaker.Staker.SequenceNum)
	require.Equal(t, "", aliceStaker.Staker.Memo)
}

func TestStakeNoFunds(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 20)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(0)
	require.Error(t, err)
}

func TestStakeWithMemo(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 20)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.StakeWithMemo(10, "test memo")
	require.NoError(t, err)

	aliceStaker, err := alice.GetStaker()
	require.NoError(t, err)
	require.Equal(t, alice.PublicKeyHex(), aliceStaker.Staker.PublicKey)
	require.Equal(t, f.SedaToAseda(10), aliceStaker.Staker.Staked)
	require.Equal(t, f.SedaToAseda(0), aliceStaker.Staker.PendingWithdrawal)
	require.Equal(t, uint64(1), aliceStaker.Staker.SequenceNum)
	memo, err := base64.StdEncoding.DecodeString(aliceStaker.Staker.Memo)
	require.NoError(t, err)
	require.Equal(t, "test memo", string(memo))
}

func TestUnstake(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 20)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	_, err = alice.Unstake()
	require.NoError(t, err)

	aliceStaker, err := alice.GetStaker()
	require.NoError(t, err)
	require.Equal(t, alice.PublicKeyHex(), aliceStaker.Staker.PublicKey)
	require.Equal(t, f.SedaToAseda(0), aliceStaker.Staker.Staked)
	require.Equal(t, f.SedaToAseda(10), aliceStaker.Staker.PendingWithdrawal)
	require.Equal(t, uint64(2), aliceStaker.Staker.SequenceNum)
	require.Equal(t, "", aliceStaker.Staker.Memo)

	// double calling unstake shouldn't matter
	_, err = alice.Unstake()
	require.NoError(t, err)
}

func TestWithdrawSelf(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 20)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	// TODO: need to reward them

	_, err = alice.Withdraw(nil)
	require.NoError(t, err)

	_, err = alice.GetStaker()
	require.Error(t, err)

	// aliceBalance := alice.Balance()
	// require.Equal(t, f.SedaToAseda(20), aliceBalance)
}

func TestWithdrawToAnother(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 20)
	bob := f.CreateTestAccount("bob", 20)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	// TODO: reward alice instead
	_, err = alice.Unstake()
	require.NoError(t, err)

	_, err = alice.Withdraw(&bob)
	require.NoError(t, err)

	_, err = alice.GetStaker()
	require.Error(t, err)

	bobBalance := bob.Balance()
	require.Equal(t, f.SedaToAseda(30), bobBalance)
}

func TestWithdrawRemovesStakerIfNoStake(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	_, err = alice.Unstake()
	require.NoError(t, err)

	_, err = alice.Withdraw(nil)
	require.NoError(t, err)

	aliceStaker, err := alice.GetStaker()
	require.Nil(t, aliceStaker.Staker)
	require.Error(t, err)
}

func TestCannotStakeIfPaused(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	f.Creator.Pause()

	_, err := alice.Stake(10)
	require.Error(t, err)
}

func TestCannotUnstakeIfPaused(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	f.Creator.Pause()

	_, err = alice.Unstake()
	require.Error(t, err)
}

func TestCannotWithdrawIfPaused(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)
	f.Creator.AddToAllowlist(alice.PublicKeyHex())

	_, err := alice.Stake(10)
	require.NoError(t, err)

	_, err = alice.Unstake()
	require.NoError(t, err)

	f.Creator.Pause()

	_, err = alice.Withdraw(nil)
	require.Error(t, err)
}
