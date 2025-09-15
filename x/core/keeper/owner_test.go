package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/stretchr/testify/require"
)

func TestGetOwner(t *testing.T) {
	f := testutil.InitFixture(t)

	someone := f.CreateTestAccount("someone", 10_000)

	resp, err := someone.GetOwner()
	require.NoError(t, err)
	require.Equal(t, f.Creator.Address(), resp.Owner)
}

func TestGetNoPendingOwner(t *testing.T) {
	f := testutil.InitFixture(t)

	someone := f.CreateTestAccount("someone", 10_000)

	resp, err := someone.GetPendingOwner()
	require.Empty(t, resp.PendingOwner)
	require.NoError(t, err)
}

func TestNonOwnerCannotTransferOwnership(t *testing.T) {
	f := testutil.InitFixture(t)

	someone := f.CreateTestAccount("someone", 10_000)

	_, err := someone.TransferOwnership("newowner")
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}

func TestTwoStepOwnershipTransfer(t *testing.T) {
	f := testutil.InitFixture(t)

	newOwner := f.CreateTestAccount("newowner", 10_000)

	// Owner transfers ownership to newOwner
	_, err := f.Creator.TransferOwnership(newOwner.Address())
	require.NoError(t, err)

	// Check pending owner is newOwner
	resp, err := newOwner.GetPendingOwner()
	require.NoError(t, err)
	require.Equal(t, newOwner.Address(), resp.PendingOwner)

	// new Owner accepts ownership
	_, err = newOwner.AcceptOwnership()
	require.NoError(t, err)

	// Check owner is now newOwner
	ownerResp, err := newOwner.GetOwner()
	require.NoError(t, err)
	require.Equal(t, newOwner.Address(), ownerResp.Owner)

	// Check no pending owner
	resp, err = f.Creator.GetPendingOwner()
	require.Empty(t, resp.PendingOwner)
	require.NoError(t, err)
}

func TestNonPendingOwnerCannotAcceptOwnership(t *testing.T) {
	f := testutil.InitFixture(t)

	newOwner := f.CreateTestAccount("newowner", 10_000)
	someone := f.CreateTestAccount("someone", 10_000)

	// Owner transfers ownership to newOwner
	_, err := f.Creator.TransferOwnership(newOwner.Address())
	require.NoError(t, err)

	// someone tries to accept ownership
	_, err = someone.AcceptOwnership()
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// Check owner is still original owner
	ownerResp, err := f.Creator.GetOwner()
	require.NoError(t, err)
	require.Equal(t, f.Creator.Address(), ownerResp.Owner)

	// Check pending owner is still newOwner
	pendingResp, err := newOwner.GetPendingOwner()
	require.NoError(t, err)
	require.Equal(t, newOwner.Address(), pendingResp.PendingOwner)
}
