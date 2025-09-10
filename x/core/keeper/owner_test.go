package keeper_test

import (
	"testing"

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
	require.NoError(t, err)
	require.Equal(t, f.Creator.Address(), resp.PendingOwner)
}

// TODO: transfer ownership test
// func TestNonOwnerCannotTransferOwnership(t *testing.T) {
// 	f := testutil.InitFixture(t)

// 	someone := f.CreateTestAccount("someone", 10_000)

// 	resp, err := someone.GetOwner()
// 	require.NoError(t, err)
// 	require.Equal(t, f.Creator.Address(), resp.Owner)
// }

// TODO: test two step ownership transfer
// TODO: test non transferee cannot accept ownership
