package keeper_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func TestUpdateDataRequestIndexing(t *testing.T) {
	f := testutil.InitFixture(t)
	f.AddStakers(t, 1)

	// Create test data request
	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())

	// Post a data request to get it in COMMITTING status
	postResult := f.PostDataRequest(
		execProgram.Hash,
		tallyProgram.Hash,
		base64.StdEncoding.EncodeToString([]byte("test-input")),
		1,
	)

	// Get the data request
	dr, err := f.CoreKeeper.GetDataRequest(f.Context(), postResult.DrID)
	require.NoError(t, err)
	require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, dr.Status)

	newStatusCommitting := types.DATA_REQUEST_STATUS_COMMITTING
	newStatusRevealing := types.DATA_REQUEST_STATUS_REVEALING
	newStatusTallying := types.DATA_REQUEST_STATUS_TALLYING

	t.Run("transition to COMMITTING status (new data request)", func(t *testing.T) {
		// Create a new data request for this test
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-2")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, newDr.Status)

		drs, _, _, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_COMMITTING, 100, nil)
		require.NoError(t, err)
		require.Contains(t, drs, newDr)
	})

	t.Run("transition to REVEALING status from COMMITTING", func(t *testing.T) {
		// Use the original data request that's in COMMITTING status
		drCopy := dr

		// This should succeed as the data request is in COMMITTING status
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &drCopy, &newStatusRevealing)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, drCopy.Status)

		// Verify the data request was moved from COMMITTING to REVEALING in the index
		drs, _, _, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_COMMITTING, 100, nil)
		require.NoError(t, err)
		require.NotContains(t, drs, drCopy)

		drs, _, _, err = f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_REVEALING, 100, nil)
		require.NoError(t, err)
		require.Contains(t, drs, drCopy)
	})

	t.Run("transition to TALLYING status from COMMITTING (timeout case)", func(t *testing.T) {
		// Create a new data request for this test
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-3")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, newDr.Status)

		// This should succeed as COMMITTING -> TALLYING is allowed for timeout cases
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusTallying)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, newDr.Status)

		// Verify the data request was moved from COMMITTING to TALLYING in the index
		drs, _, _, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_COMMITTING, 100, nil)
		require.NoError(t, err)
		require.NotContains(t, drs, newDr)

		drs, _, _, err = f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_TALLYING, 100, nil)
		require.NoError(t, err)
		require.Contains(t, drs, newDr)
	})

	t.Run("transition to TALLYING status from REVEALING", func(t *testing.T) {
		// Create a new data request and transition it to REVEALING first
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-4")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// First transition to REVEALING
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusRevealing)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, newDr.Status)

		// Now transition to TALLYING
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusTallying)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, newDr.Status)

		// Verify the data request was moved from REVEALING to TALLYING in the index
		drs, _, _, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_REVEALING, 100, nil)
		require.NoError(t, err)
		require.NotContains(t, drs, newDr)

		drs, _, _, err = f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_TALLYING, 100, nil)
		require.NoError(t, err)
		require.Contains(t, drs, newDr)
	})

	t.Run("invalid transition to UNSPECIFIED status", func(t *testing.T) {
		// Create a new data request for this test
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-5")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// This should fail as UNSPECIFIED is not a valid target status
		newStatusUnspecified := types.DATA_REQUEST_STATUS_UNSPECIFIED
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusUnspecified)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStatusTransition)
	})

	t.Run("invalid status transition from TALLYING to REVEALING", func(t *testing.T) {
		// Create a new data request and transition it to TALLYING
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-7")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// Transition to TALLYING
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusTallying)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, newDr.Status)

		// Now try to transition back to REVEALING - this should fail
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusRevealing)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStatusTransition)
	})

	t.Run("invalid status transition from REVEALING to COMMITTING", func(t *testing.T) {
		// Create a new data request and transition it to REVEALING
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-8")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// Transition to REVEALING
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusRevealing)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, newDr.Status)

		// Now try to transition back to COMMITTING - this should fail
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusCommitting)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStatusTransition)
	})

	t.Run("error when data request is not found", func(t *testing.T) {
		// Create a new data request for this test
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-9")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// Remove the data request
		err = f.CoreKeeper.RemoveDataRequest(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		require.NoError(t, err)

		// Now try to transition to TALLYING - this should fail
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusTallying)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrDataRequestNotFound)
	})
}
