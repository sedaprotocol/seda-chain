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
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 1)

	// Create test data request
	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.HTTPHeavyWasm(), f.Context().BlockTime())
	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())
	dr := testutil.NewTestDR(
		execProgram.Hash,
		tallyProgram.Hash,
		[]byte("reveal"),
		base64.StdEncoding.EncodeToString([]byte("test-input")),
		150000000000000000,
		0,
		[]string{},
		1,
	)

	// Post a data request to get it in COMMITTING status
	dr.PostDataRequest(f)

	// Get the data request
	checkDR, err := f.CoreKeeper.GetDataRequest(f.Context(), dr.GetDataRequestID())
	require.NoError(t, err)
	require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, checkDR.Status)

	newStatusCommitting := types.DATA_REQUEST_STATUS_COMMITTING
	newStatusRevealing := types.DATA_REQUEST_STATUS_REVEALING
	newStatusTallying := types.DATA_REQUEST_STATUS_TALLYING

	t.Run("transition to COMMITTING status (new data request)", func(t *testing.T) {
		dr2 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-2")),
			150000000000000000,
			0,
			[]string{},
			1,
		)

		// Create a new data request for this test
		dr2.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr2.GetDataRequestID())
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, newDr.Status)

		drs, _, _, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_COMMITTING, 100, nil)
		require.NoError(t, err)
		require.Contains(t, drs, newDr)
	})

	t.Run("transition to REVEALING status from COMMITTING", func(t *testing.T) {
		// Use the original data request that's in COMMITTING status

		// This should succeed as the data request is in COMMITTING status
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &checkDR, &newStatusRevealing)
		require.NoError(t, err)
		require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, checkDR.Status)

		// Verify the data request was moved from COMMITTING to REVEALING in the index
		drs, _, _, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_COMMITTING, 100, nil)
		require.NoError(t, err)
		require.NotContains(t, drs, checkDR)

		drs, _, _, err = f.CoreKeeper.GetDataRequestsByStatus(f.Context(), types.DATA_REQUEST_STATUS_REVEALING, 100, nil)
		require.NoError(t, err)
		require.Contains(t, drs, checkDR)
	})

	t.Run("transition to TALLYING status from COMMITTING (timeout case)", func(t *testing.T) {
		dr3 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-3")),
			150000000000000000,
			0,
			[]string{},
			1,
		)

		dr3.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr3.GetDataRequestID())
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
		dr4 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-4")),
			150000000000000000,
			0,
			[]string{},
			1,
		)

		dr4.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr4.GetDataRequestID())
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
		dr5 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-5")),
			150000000000000000,
			0,
			[]string{},
			1,
		)

		dr5.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr5.GetDataRequestID())
		require.NoError(t, err)

		// This should fail as UNSPECIFIED is not a valid target status
		newStatusUnspecified := types.DATA_REQUEST_STATUS_UNSPECIFIED
		err = f.CoreKeeper.UpdateDataRequest(f.Context(), &newDr, &newStatusUnspecified)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidStatusTransition)
	})

	t.Run("invalid status transition from TALLYING to REVEALING", func(t *testing.T) {
		// Create a new data request and transition it to TALLYING
		dr6 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-6")),
			150000000000000000,
			0,
			[]string{},
			1,
		)
		dr6.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr6.GetDataRequestID())
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
		dr7 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-7")),
			150000000000000000,
			0,
			[]string{},
			1,
		)
		dr7.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr7.GetDataRequestID())
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
		dr8 := testutil.NewTestDR(
			execProgram.Hash,
			tallyProgram.Hash,
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte("test-input-8")),
			150000000000000000,
			0,
			[]string{},
			1,
		)
		dr8.PostDataRequest(f)

		newDr, err := f.CoreKeeper.GetDataRequest(f.Context(), dr8.GetDataRequestID())
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
