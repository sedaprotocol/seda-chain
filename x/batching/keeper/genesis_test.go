package keeper_test

import (
	"crypto/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func TestExportGenesis(t *testing.T) {
	f := initFixture(t)

	f.pubKeyKeeper.SetProvingScheme(f.Context(), pubkeytypes.ProvingScheme{
		Index:       uint32(sedatypes.SEDAKeyIndexSecp256k1),
		IsActivated: true,
	})

	valAddrs, _, _ := addBatchSigningValidators(t, f, 10)

	dataResults := generateDataResults(t, 25)
	for _, dr := range dataResults {
		err := f.batchingKeeper.SetDataResultForBatching(f.Context(), dr)
		require.NoError(t, err)
	}

	err := f.batchingKeeper.EndBlock(f.Context())
	require.NoError(t, err)

	for _, valAddr := range valAddrs {
		randSigBytes := make([]byte, 10)
		rand.Read(randSigBytes)
		err = f.batchingKeeper.SetBatchSigSecp256k1(f.Context(), collections.DefaultSequenceStart+1, sdk.ValAddress(valAddr), randSigBytes)
		require.NoError(t, err)
	}

	curBatchNumBefore, err := f.batchingKeeper.GetCurrentBatchNum(f.Context())
	require.NoError(t, err)

	latestBatchBefore, err := f.batchingKeeper.GetLatestBatch(f.Context())
	require.NoError(t, err)

	batchDataBefore, err := f.batchingKeeper.GetBatchData(f.Context(), latestBatchBefore.BatchNumber)
	require.NoError(t, err)

	batchSigsBefore, err := f.batchingKeeper.GetBatchSignatures(f.Context(), latestBatchBefore.BatchNumber)
	require.NoError(t, err)

	// Export and import genesis.
	exportGenesis := f.batchingKeeper.ExportGenesis(f.Context())

	err = types.ValidateGenesis(exportGenesis)
	require.NoError(t, err)

	f.batchingKeeper.InitGenesis(f.Context(), exportGenesis)

	// Compare imported state against the state before export.
	curBatchNumAfter, err := f.batchingKeeper.GetCurrentBatchNum(f.Context())
	require.NoError(t, err)
	require.Equal(t, curBatchNumBefore, curBatchNumAfter)

	batchesAfter, err := f.batchingKeeper.GetAllBatches(f.Context())
	require.NoError(t, err)
	require.Equal(t, []types.Batch{latestBatchBefore}, batchesAfter)

	batchDataAfter, err := f.batchingKeeper.GetBatchData(f.Context(), latestBatchBefore.BatchNumber)
	require.NoError(t, err)
	require.Equal(t, batchDataBefore, batchDataAfter)

	dataResultsAfter, err := f.batchingKeeper.GetDataResults(f.Context(), true)
	require.NoError(t, err)
	require.ElementsMatch(t, dataResults, dataResultsAfter)

	for _, dataResults := range dataResults {
		batchAssignment, err := f.batchingKeeper.GetBatchAssignment(f.Context(), dataResults.DrId, dataResults.DrBlockHeight)
		require.NoError(t, err)
		require.Equal(t, batchAssignment, latestBatchBefore.BatchNumber)
	}

	batchSigsAfter, err := f.batchingKeeper.GetBatchSignatures(f.Context(), latestBatchBefore.BatchNumber)
	require.NoError(t, err)
	require.ElementsMatch(t, batchSigsBefore, batchSigsAfter)
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	gs := types.DefaultGenesisState()
	err := types.ValidateGenesis(*gs)
	require.NoError(suite.T(), err)

	keeper := suite.keeper
	keeper.InitGenesis(suite.ctx, *gs)

	curBatchNum, err := keeper.GetCurrentBatchNum(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(collections.DefaultSequenceStart, curBatchNum)
}
