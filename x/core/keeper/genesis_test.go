package keeper_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// Test export and import genesis after simulated data request flows.
func TestExportImport(t *testing.T) {
	// Setup with arbitrary data
	f := testutil.InitFixture(t, false, nil)
	f.AddStakers(t, 32)

	gs := types.DefaultGenesisState()
	gs.Params.TallyConfig.FilterGasCostNone = 200_000
	gs.Params.StakingConfig.MinimumStake = math.NewInt(500000000000000000)
	gs.Params.DataRequestConfig.MemoLimitInBytes = 150
	f.CoreKeeper.SetParams(f.Context(), gs.Params)

	totalDRs := 123
	drIDs := make([]string, totalDRs)
	rf := 3
	var postedDRs, committedDRs, revealedDRs int
	for j := 0; j < totalDRs; j++ {
		// Decide whether to post, post + commit, or post + commit + reveal.
		var numCommits, numReveals int
		switch j % 3 {
		case 0:
			postedDRs++
			numCommits = rand.Intn(rf)
		case 1:
			committedDRs++
			numCommits = rf
			numReveals = rand.Intn(rf)
		case 2:
			revealedDRs++
			numCommits = rf
			numReveals = rf
		}

		dr := testutil.NewTestDRWithRandomPrograms(
			[]byte("reveal"),
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
			150000000000000000,
			0,
			[]string{},
			rf,
			f.Context().BlockTime(),
		)

		drID := dr.ExecuteDataRequestFlow(f, numCommits, numReveals, false)

		drIDs[j] = drID
	}

	err := f.CoreKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	f.AddBlock()

	// Retrieve the state before export.
	ownerBeforeExport, err := f.CoreKeeper.GetOwner(f.Context())
	require.NoError(t, err)

	pendingOwnerBeforeExport, err := f.CoreKeeper.GetPendingOwner(f.Context())
	require.NoError(t, err)

	pausedBeforeExport, err := f.CoreKeeper.IsPaused(f.Context())
	require.NoError(t, err)

	paramsBeforeExport, err := f.CoreKeeper.GetParams(f.Context())
	require.NoError(t, err)

	var drsBeforeExport []types.DataRequest
	for _, drID := range drIDs {
		dr, err := f.CoreKeeper.GetDataRequest(f.Context(), drID)
		if err == nil {
			drsBeforeExport = append(drsBeforeExport, dr)
		}
	}

	var allowlistBeforeExport []string
	allowlistBeforeExport, err = f.CoreKeeper.ListAllowlist(f.Context())
	require.NoError(t, err)

	var stakersBeforeExport []types.Staker
	stakersBeforeExport, err = f.CoreKeeper.GetAllStakers(f.Context())
	require.NoError(t, err)

	var revealBodiesBeforeExport []types.RevealBody
	revealBodiesBeforeExport, err = f.CoreKeeper.GetAllRevealBodies(f.Context())
	require.NoError(t, err)

	unspecifiedBeforeExport, err := f.CoreKeeper.GetDataRequestIDsByStatus(f.Context(), types.DATA_REQUEST_STATUS_UNSPECIFIED)
	require.NoError(t, err)
	committingBeforeExport, err := f.CoreKeeper.GetDataRequestIDsByStatus(f.Context(), types.DATA_REQUEST_STATUS_COMMITTING)
	require.NoError(t, err)
	revealingBeforeExport, err := f.CoreKeeper.GetDataRequestIDsByStatus(f.Context(), types.DATA_REQUEST_STATUS_REVEALING)
	require.NoError(t, err)
	tallyingBeforeExport, err := f.CoreKeeper.GetDataRequestIDsByStatus(f.Context(), types.DATA_REQUEST_STATUS_TALLYING)
	require.NoError(t, err)

	timeoutQueueBeforeExport, err := f.CoreKeeper.ListTimeoutQueue(f.Context())
	require.NoError(t, err)
	require.Equal(t, postedDRs+committedDRs, len(timeoutQueueBeforeExport))

	// Export and import genesis.
	exportedGenesis := f.CoreKeeper.ExportGenesis(f.Context())

	f2 := testutil.InitFixture(t, false, nil)

	err = types.ValidateGenesis(exportedGenesis)
	require.NoError(t, err)

	f2.CoreKeeper.InitGenesis(f2.Context(), exportedGenesis)

	// Compare the imported state against the state before export.
	ownerAfterImport, err := f2.CoreKeeper.GetOwner(f2.Context())
	require.NoError(t, err)
	require.Equal(t, ownerBeforeExport, ownerAfterImport)

	pendingOwnerAfterImport, err := f2.CoreKeeper.GetPendingOwner(f2.Context())
	require.NoError(t, err)
	require.Equal(t, pendingOwnerBeforeExport, pendingOwnerAfterImport)

	pausedAfterImport, err := f2.CoreKeeper.IsPaused(f2.Context())
	require.NoError(t, err)
	require.Equal(t, pausedBeforeExport, pausedAfterImport)

	paramsAfterImport, err := f2.CoreKeeper.GetParams(f2.Context())
	require.NoError(t, err)
	require.Equal(t, paramsBeforeExport, paramsAfterImport)

	drsAfterImport, err := f2.CoreKeeper.GetAllDataRequests(f2.Context())
	require.NoError(t, err)
	require.Equal(t, postedDRs+committedDRs, len(drsAfterImport)) // revealed DRs are removed from data request store
	require.ElementsMatch(t, drsBeforeExport, drsAfterImport)

	for _, dr := range drsAfterImport {
		if len(dr.Commits) < int(dr.ReplicationFactor) {
			require.Equal(t, types.DATA_REQUEST_STATUS_COMMITTING, dr.Status)
		} else if len(dr.Reveals) < int(dr.ReplicationFactor) {
			require.Equal(t, types.DATA_REQUEST_STATUS_REVEALING, dr.Status)
		} else {
			// This is unreachable because tally endblock execution has removed
			// tallying data requests from the store.
			require.Equal(t, types.DATA_REQUEST_STATUS_TALLYING, dr.Status)
		}

		for executor := range dr.Reveals {
			revealBody, err := f2.CoreKeeper.GetRevealBody(f2.Context(), dr.ID, executor)
			require.NoError(t, err)
			require.Equal(t, dr.ID, revealBody.DrID)
			require.Equal(t, executor, executor)
		}
	}

	allowlistAfterImport, err := f2.CoreKeeper.ListAllowlist(f2.Context())
	require.NoError(t, err)
	require.Equal(t, allowlistBeforeExport, allowlistAfterImport)

	stakersAfterImport, err := f2.CoreKeeper.GetAllStakers(f2.Context())
	require.NoError(t, err)
	require.Equal(t, stakersBeforeExport, stakersAfterImport)

	revealBodiesAfterImport, err := f2.CoreKeeper.GetAllRevealBodies(f2.Context())
	require.NoError(t, err)
	require.Equal(t, revealBodiesBeforeExport, revealBodiesAfterImport)

	unspecifiedAfterImport, err := f2.CoreKeeper.GetDataRequestIDsByStatus(f2.Context(), types.DATA_REQUEST_STATUS_UNSPECIFIED)
	require.NoError(t, err)
	require.Equal(t, unspecifiedBeforeExport, unspecifiedAfterImport)
	committingAfterImport, err := f2.CoreKeeper.GetDataRequestIDsByStatus(f2.Context(), types.DATA_REQUEST_STATUS_COMMITTING)
	require.NoError(t, err)
	require.Equal(t, committingBeforeExport, committingAfterImport)
	revealingAfterImport, err := f2.CoreKeeper.GetDataRequestIDsByStatus(f2.Context(), types.DATA_REQUEST_STATUS_REVEALING)
	require.NoError(t, err)
	require.Equal(t, revealingBeforeExport, revealingAfterImport)
	tallyingAfterImport, err := f2.CoreKeeper.GetDataRequestIDsByStatus(f2.Context(), types.DATA_REQUEST_STATUS_TALLYING)
	require.NoError(t, err)
	require.Equal(t, tallyingBeforeExport, tallyingAfterImport)

	timeoutQueueAfterImport, err := f2.CoreKeeper.ListTimeoutQueue(f2.Context())
	require.NoError(t, err)
	require.Equal(t, timeoutQueueBeforeExport, timeoutQueueAfterImport)

	// Second export and comparison of exported states
	exportedGenesis2 := f2.CoreKeeper.ExportGenesis(f2.Context())
	err = types.ValidateGenesis(exportedGenesis2)
	require.NoError(t, err)
	require.Equal(t, exportedGenesis, exportedGenesis2)
}
