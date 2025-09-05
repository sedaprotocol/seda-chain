package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func TestExportGenesis(t *testing.T) {
	f := initFixture(t)

	// TODO Test rest of the params
	gs := types.DefaultGenesisState()
	gs.Params.TallyConfig.FilterGasCostNone = 200_000
	gs.Params.TallyConfig.FilterGasCostMultiplierMode = 400_000
	gs.Params.TallyConfig.FilterGasCostMultiplierMAD = 600_000
	f.keeper.SetParams(f.Context(), gs.Params)

	err := types.ValidateGenesis(*gs)
	require.NoError(t, err)

	// Export and import genesis.
	exportGenesis := f.keeper.ExportGenesis(f.Context())

	err = types.ValidateGenesis(exportGenesis)
	require.NoError(t, err)

	f.keeper.InitGenesis(f.Context(), exportGenesis)

	afterParams, err := f.keeper.GetParams(f.Context())
	require.NoError(t, err)
	require.Equal(t, gs.Params, afterParams)
}
