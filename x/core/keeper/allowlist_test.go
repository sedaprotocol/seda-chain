package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestAddToAllowlistUnauthorized(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	_, err := alice.Stake(10)
	require.ErrorIs(t, err, types.ErrNotAllowlisted)
}

func TestOnlyOwnerCanAddToAllowlist(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)
	bob := f.CreateTestAccount("bob", 10_000)

	_, err := alice.AddToAllowlist(bob.PublicKeyHex())
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}

func TestOnlyOwnerCanRemoveFromAllowlist(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	// owner can add to allowlist
	_, err := f.Creator.AddToAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// non-owner cannot remove from allowlist
	_, err = alice.RemoveFromAllowlist(alice.PublicKeyHex())
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// owner can remove from allowlist
	_, err = f.Creator.RemoveFromAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)
}

func TestAddToAllowlistWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	// owner can add to allowlist
	_, err := f.Creator.AddToAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// alice can now stake
	_, err = alice.Stake(10)
	t.Log("stake err", err)
	require.NoError(t, err)

	// alice can unstake and withdraw
	_, err = alice.Unstake()
	require.NoError(t, err)
	_, err = alice.Withdraw(&alice)
	require.NoError(t, err)

	// owner can remove from allowlist
	_, err = f.Creator.RemoveFromAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// alice can no longer stake
	_, err = alice.Stake(10)
	require.ErrorIs(t, err, types.ErrNotAllowlisted)

	// update the config to disable allowlist
	_, err = f.Creator.SetStakingConfig(types.StakingConfig{
		AllowlistEnabled: false,
		MinimumStake:     math.NewInt(10),
	})
	require.NoError(t, err)

	// alice can now stake
	_, err = alice.Stake(10)
	require.NoError(t, err)
}

func TestGetAllowlist(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// initially empty and called by owner
	resp, err := f.Creator.GetAllowlist()
	require.NoError(t, err)
	require.Len(t, resp.PublicKeys, 0)

	// Add alice to allowlist
	alice := f.CreateTestAccount("alice", 10_000)
	_, err = f.Creator.AddToAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// Check allowlist has alice and called by alice
	resp, err = alice.GetAllowlist()
	require.NoError(t, err)
	require.Len(t, resp.PublicKeys, 1)
	require.Equal(t, alice.PublicKeyHex(), resp.PublicKeys[0])

	// Add bob to allowlist
	bob := f.CreateTestAccount("bob", 10_000)
	_, err = f.Creator.AddToAllowlist(bob.PublicKeyHex())
	require.NoError(t, err)

	// Check allowlist has alice and bob and called by bob
	resp, err = bob.GetAllowlist()
	require.NoError(t, err)
	require.Len(t, resp.PublicKeys, 2)
	// check the array contains both keys - order not guaranteed
	require.Contains(t, resp.PublicKeys, alice.PublicKeyHex())
	require.Contains(t, resp.PublicKeys, bob.PublicKeyHex())
}

func TestRemoveFromAllowlistUnstakesUser(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	// owner can add to allowlist
	_, err := f.Creator.AddToAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// alice can now stake
	_, err = alice.Stake(10)
	require.NoError(t, err)

	// owner can remove from allowlist
	_, err = f.Creator.RemoveFromAllowlist(alice.PublicKeyHex())
	require.NoError(t, err)

	// alice should now be unstaked
	stakeResp, err := alice.GetStaker()
	require.NoError(t, err)
	require.Equal(t, math.ZeroInt(), stakeResp.Staker.Staked)
	require.Equal(t, f.SedaToAseda(10), stakeResp.Staker.PendingWithdrawal)
}
