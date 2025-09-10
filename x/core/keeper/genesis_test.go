package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func TestExportGenesis(t *testing.T) {
	f := testutil.InitFixture(t)

	// TODO Test rest of the params
	gs := types.DefaultGenesisState()
	gs.Params.TallyConfig.FilterGasCostNone = 200_000
	gs.Params.TallyConfig.FilterGasCostMultiplierMode = 400_000
	gs.Params.TallyConfig.FilterGasCostMultiplierMAD = 600_000
	f.CoreKeeper.SetParams(f.Context(), gs.Params)

	err := types.ValidateGenesis(*gs)
	require.NoError(t, err)

	// Export and import genesis.
	exportGenesis := f.CoreKeeper.ExportGenesis(f.Context())

	err = types.ValidateGenesis(exportGenesis)
	require.NoError(t, err)

	f.CoreKeeper.InitGenesis(f.Context(), exportGenesis)

	afterParams, err := f.CoreKeeper.GetParams(f.Context())
	require.NoError(t, err)
	require.Equal(t, gs.Params, afterParams)
}
