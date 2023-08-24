package storage_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/sedaprotocol/seda-chain/testutil/keeper"
	"github.com/sedaprotocol/seda-chain/testutil/nullify"
	storage "github.com/sedaprotocol/seda-chain/x/storage"
	"github.com/sedaprotocol/seda-chain/x/storage/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SedachainKeeper(t)
	err := storage.InitGenesis(ctx, *k, genesisState)
	require.NoError(t, err)

	got := storage.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
