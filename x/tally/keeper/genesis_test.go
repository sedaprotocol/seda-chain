package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

func TestExportGenesis(t *testing.T) {
	f := initFixture(t)

	gs := types.DefaultGenesisState()
	gs.Params.FilterGasCostNone = 200_000
	gs.Params.FilterGasCostMultiplierMode = 400_000
	gs.Params.FilterGasCostMultiplierMAD = 600_000
	f.tallyKeeper.SetParams(f.Context(), gs.Params)

	err := types.ValidateGenesis(*gs)
	require.NoError(t, err)

	// Export and import genesis.
	exportGenesis := f.tallyKeeper.ExportGenesis(f.Context())

	err = types.ValidateGenesis(exportGenesis)
	require.NoError(t, err)

	f.tallyKeeper.InitGenesis(f.Context(), exportGenesis)

	afterParams, err := f.tallyKeeper.GetParams(f.Context())
	require.NoError(t, err)
	require.Equal(t, gs.Params, afterParams)
}
