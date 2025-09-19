package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func TestStakerIndexing(t *testing.T) {
	f := testutil.InitFixture(t)

	// Create test stakers
	staker1 := types.Staker{
		PublicKey:         "pubkey1",
		Memo:              "test staker 1",
		Staked:            math.NewInt(1000),
		PendingWithdrawal: math.NewInt(0),
	}

	staker2 := types.Staker{
		PublicKey:         "pubkey2",
		Memo:              "test staker 2",
		Staked:            math.NewInt(2000),
		PendingWithdrawal: math.NewInt(100),
	}

	staker3 := types.Staker{
		PublicKey:         "pubkey3",
		Memo:              "test staker 3",
		Staked:            math.NewInt(3000),
		PendingWithdrawal: math.NewInt(0),
	}

	t.Run("SetStaker method - adding new stakers", func(t *testing.T) {
		// Test adding first staker
		err := f.CoreKeeper.SetStaker(f.Context(), staker1)
		require.NoError(t, err)

		// Verify count is updated
		count, err := f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(1), count)

		// Test adding second staker
		err = f.CoreKeeper.SetStaker(f.Context(), staker2)
		require.NoError(t, err)

		// Verify count is updated
		count, err = f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(2), count)

		// Test adding third staker
		err = f.CoreKeeper.SetStaker(f.Context(), staker3)
		require.NoError(t, err)

		// Verify count is updated
		count, err = f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(3), count)
	})

	t.Run("GetStaker method - retrieving existing stakers", func(t *testing.T) {
		// Get first staker
		retrievedStaker1, err := f.CoreKeeper.GetStaker(f.Context(), "pubkey1")
		require.NoError(t, err)
		require.Equal(t, staker1, retrievedStaker1)

		// Get second staker
		retrievedStaker2, err := f.CoreKeeper.GetStaker(f.Context(), "pubkey2")
		require.NoError(t, err)
		require.Equal(t, staker2, retrievedStaker2)

		// Get third staker
		retrievedStaker3, err := f.CoreKeeper.GetStaker(f.Context(), "pubkey3")
		require.NoError(t, err)
		require.Equal(t, staker3, retrievedStaker3)
	})

	t.Run("GetStaker method - error when staker does not exist", func(t *testing.T) {
		// Try to get non-existing staker
		_, err := f.CoreKeeper.GetStaker(f.Context(), "nonexistent")
		require.Error(t, err)
	})

	t.Run("GetStakerCount method", func(t *testing.T) {
		count, err := f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(3), count)
	})

	t.Run("RemoveStaker method - removing last staker", func(t *testing.T) {
		// Remove the last staker (pubkey3, index 2)
		err := f.CoreKeeper.RemoveStaker(f.Context(), "pubkey3")
		require.NoError(t, err)

		// Verify count decreased
		count, err := f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(2), count)

		// Verify staker is removed
		_, err = f.CoreKeeper.GetStaker(f.Context(), "pubkey3")
		require.Error(t, err)

		// Verify remaining stakers are still accessible
		_, err = f.CoreKeeper.GetStaker(f.Context(), "pubkey1")
		require.NoError(t, err)
		_, err = f.CoreKeeper.GetStaker(f.Context(), "pubkey2")
		require.NoError(t, err)

		// Verify that the indices have not changed.
		f.CheckStakerIndexing(t, "pubkey1", 0, true)
		f.CheckStakerIndexing(t, "pubkey2", 1, true)
		f.CheckStakerIndexing(t, "pubkey3", 2, false)
	})

	t.Run("RemoveStaker method - removing middle staker (swap logic)", func(t *testing.T) {
		// Add back a third staker to test swap logic
		staker4 := types.Staker{
			PublicKey:         "pubkey4",
			Memo:              "test staker 4",
			Staked:            math.NewInt(4000),
			PendingWithdrawal: math.NewInt(0),
		}
		err := f.CoreKeeper.SetStaker(f.Context(), staker4)
		require.NoError(t, err)

		// Verify we have 3 stakers now
		count, err := f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(3), count)

		// Remove the first staker (pubkey1, index 0)
		// This should trigger the swap logic where the last staker (pubkey4, index 2)
		// gets moved to index 0
		err = f.CoreKeeper.RemoveStaker(f.Context(), "pubkey1")
		require.NoError(t, err)

		// Verify count decreased
		count, err = f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(2), count)

		// Verify pubkey1 is removed
		_, err = f.CoreKeeper.GetStaker(f.Context(), "pubkey1")
		require.Error(t, err)

		// Verify remaining stakers are still accessible
		_, err = f.CoreKeeper.GetStaker(f.Context(), "pubkey2")
		require.NoError(t, err)
		_, err = f.CoreKeeper.GetStaker(f.Context(), "pubkey4")
		require.NoError(t, err)

		// Verify that the item at last index has been swapped to the index of
		// the removed item.
		f.CheckStakerIndexing(t, "pubkey4", 0, true)
		f.CheckStakerIndexing(t, "pubkey2", 1, true)
		f.CheckStakerIndexing(t, "pubkey1", 2, false)
	})

	t.Run("RemoveStaker method - removing only staker", func(t *testing.T) {
		// Remove remaining stakers to test single staker removal
		err := f.CoreKeeper.RemoveStaker(f.Context(), "pubkey2")
		require.NoError(t, err)
		err = f.CoreKeeper.RemoveStaker(f.Context(), "pubkey4")
		require.NoError(t, err)

		// Verify count is 0
		count, err := f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(0), count)

		// Add one staker back
		staker5 := types.Staker{
			PublicKey:         "pubkey5",
			Memo:              "test staker 5",
			Staked:            math.NewInt(5000),
			PendingWithdrawal: math.NewInt(0),
		}
		err = f.CoreKeeper.SetStaker(f.Context(), staker5)
		require.NoError(t, err)

		// Remove the only staker
		err = f.CoreKeeper.RemoveStaker(f.Context(), "pubkey5")
		require.NoError(t, err)

		// Verify count is 0
		count, err = f.CoreKeeper.GetStakerCount(f.Context())
		require.NoError(t, err)
		require.Equal(t, uint32(0), count)
	})

	t.Run("RemoveStaker method - error when staker does not exist", func(t *testing.T) {
		// Try to remove non-existing staker
		err := f.CoreKeeper.RemoveStaker(f.Context(), "nonexistent")
		require.Error(t, err)
	})
}
