package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestAddToAllowlistUnauthorized(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	_, err := alice.Stake(10, "t")
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

	// _, err = alice.Stake(10, "")
	// t.Log("stake err", err)
	// require.NoError(t, err)

	// TODO: not yet implemented
	// _, err = alice.Unstake()
	// require.NoError(t, err)
	// _, err = alice.Withdraw()
	// require.NoError(t, err)

	// _, err = f.Creator.RemoveFromAllowlist(alice.PublicKeyHex())
	// require.NoError(t, err)

	// _, err = alice.Stake(10, "t")
	// require.ErrorIs(t, err, types.ErrNotAllowlisted)
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

// TODO: test removing from allowlist unstakes user

// TODO: test to update config to disable allowlist test

func TestPauseBasics(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// initially not paused
	paused, err := f.CoreKeeper.IsPaused(f.Context())
	require.NoError(t, err)
	require.False(t, paused)

	// pause the contract
	_, err = f.Creator.Pause()
	require.NoError(t, err)

	// cannot pause again
	_, err = f.Creator.Pause()
	require.Error(t, err)

	// unpause the contract
	_, err = f.Creator.Unpause()
	require.NoError(t, err)

	// cannot unpause again
	_, err = f.Creator.Unpause()
	require.Error(t, err)
}

func TestOnlyOwnerCanPause(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	// non-owner cannot pause
	_, err := alice.Pause()
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// non-owner cannot unpause
	_, err = alice.Unpause()
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}
