package keeper_test

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestAddToAllowlistUnauthorized(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)

	_, err := alice.Stake(10, "t")
	require.ErrorIs(t, err, types.ErrNotAllowlisted)
}

func TestAddToAllowlistWorks(t *testing.T) {
	f := testutil.InitFixture(t)

	alice := f.CreateTestAccount("alice", 10_000)

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

// func TestGetAllowlist(t *testing.T) {
// 	f := testutil.InitFixture(t)

// 	resp, err := f.Creator.GetAllowlist(f.Creator.PublicKeyHex())
// 	require.NoError(t, err)
// 	require.Len(t, resp.PublicKeys, 0)

// 	alice := f.CreateTestAccount("alice", 10_000)
// 	_, err = f.Creator.AddToAllowlist(alice.PublicKeyHex())
// 	require.NoError(t, err)

// 	resp, err = f.Creator.GetAllowlist(alice.PublicKeyHex())
// 	require.NoError(t, err)
// 	require.Len(t, resp.PublicKeys, 1)
// 	require.Equal(t, alice.PublicKeyHex(), resp.PublicKeys[0])

// 	bob := f.CreateTestAccount("bob", 10_000)
// 	_, err = f.Creator.AddToAllowlist(bob.PublicKeyHex())
// 	require.NoError(t, err)

// 	resp, err = f.Creator.GetAllowlist(bob.PublicKeyHex())
// 	require.NoError(t, err)
// 	require.Len(t, resp.PublicKeys, 2)
// 	require.Equal(t, alice.PublicKeyHex(), resp.PublicKeys[0])
// 	require.Equal(t, bob.PublicKeyHex(), resp.PublicKeys[1])
// }

// TODO: test removing from allowlist unstakes user

// TODO: test to update config to disable allowlist test

// TODO: test to test allow list query

// TODO: test pause works for allowlist
