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

		exists, err := f.CoreKeeper.CheckDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("transition to REVEALING status from COMMITTING", func(t *testing.T) {
		// Use the original data request that's in COMMITTING status
		drCopy := dr

		// This should succeed as the data request is in COMMITTING status
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &drCopy, types.DATA_REQUEST_STATUS_REVEALING)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, drCopy.Status)

		// Verify the data request was moved from COMMITTING to REVEALING in the index
		exists, err := f.CoreKeeper.CheckDataRequestIndexing(f.Context(), drCopy.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		require.NoError(t, err)
		require.False(t, exists)

		exists, err = f.CoreKeeper.CheckDataRequestIndexing(f.Context(), drCopy.Index(), types.DATA_REQUEST_STATUS_REVEALING)
		require.NoError(t, err)
		require.True(t, exists)
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
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_TALLYING)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, newDr.Status)

		// Verify the data request was moved from COMMITTING to TALLYING in the index
		exists, err := f.CoreKeeper.CheckDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		require.NoError(t, err)
		require.False(t, exists)

		exists, err = f.CoreKeeper.CheckDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_TALLYING)
		require.NoError(t, err)
		require.True(t, exists)
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
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_REVEALING)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, newDr.Status)

		// Now transition to TALLYING
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_TALLYING)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, newDr.Status)

		// Verify the data request was moved from REVEALING to TALLYING in the index
		exists, err := f.CoreKeeper.CheckDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_REVEALING)
		require.NoError(t, err)
		require.False(t, exists)

		exists, err = f.CoreKeeper.CheckDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_TALLYING)
		require.NoError(t, err)
		require.True(t, exists)
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
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_UNSPECIFIED)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStatusTransition)
	})

	t.Run("error when data request index not found in current status", func(t *testing.T) {
		// Create a new data request for this test
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-6")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// Manually remove the data request from the COMMITTING index to simulate it not being found
		err = f.CoreKeeper.RemoveDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		require.NoError(t, err)

		// Now try to transition to REVEALING - this should fail
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_REVEALING)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrDataRequestStatusNotFound)
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
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_TALLYING)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, newDr.Status)

		// Now try to transition back to REVEALING - this should fail
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_REVEALING)
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
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_REVEALING)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, newDr.Status)

		// Now try to transition back to COMMITTING - this should fail
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_COMMITTING)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStatusTransition)
	})

	t.Run("error when data request index not found for TALLYING transition", func(t *testing.T) {
		// Create a new data request for this test
		newPostResult := f.PostDataRequest(
			execProgram.Hash,
			tallyProgram.Hash,
			base64.StdEncoding.EncodeToString([]byte("test-input-9")),
			1,
		)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), newPostResult.DrID)
		require.NoError(t, err)

		// Manually remove the data request from the COMMITTING index to simulate it not being found
		err = f.CoreKeeper.RemoveDataRequestIndexing(f.Context(), newDr.Index(), types.DATA_REQUEST_STATUS_COMMITTING)
		require.NoError(t, err)

		// Now try to transition to TALLYING - this should fail
		err = f.CoreKeeper.UpdateDataRequestIndexing(f.Context(), &newDr, types.DATA_REQUEST_STATUS_TALLYING)
		require.Error(t, err)
		require.Contains(t, err.Error(), "data request index was not found under given status")
	})
}
