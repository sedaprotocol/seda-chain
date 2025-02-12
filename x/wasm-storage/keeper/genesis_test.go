package keeper_test

import (
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func (s *KeeperTestSuite) TestExportGenesis() {
	s.SetupTest()

	var expectedState types.GenesisState
	for i := 0; i < 10; i++ {
		oracleProgram := types.NewOracleProgram([]byte(fmt.Sprintf("test%d", i)), s.ctx.BlockTime())
		err := s.keeper.OracleProgram.Set(s.ctx, oracleProgram.Hash, oracleProgram)
		require.NoError(s.T(), err)
		expectedState.OraclePrograms = append(expectedState.OraclePrograms, oracleProgram)
	}

	coreContractAddr := "seda1egmjtl66w59kk59p55r0gspsrpysxd7xgrq5ve8qpdcratd7wxlqkj7ga5"
	s.keeper.CoreContractRegistry.Set(s.ctx, coreContractAddr)
	expectedState.CoreContractRegistry = coreContractAddr

	params := types.Params{
		MaxWasmSize:     512 * 1024,
		WasmCostPerByte: 90000000000000,
	}
	err := s.keeper.Params.Set(s.ctx, params)
	require.NoError(s.T(), err)
	expectedState.Params = params

	// Export and import genesis.
	exportedState := s.keeper.ExportGenesis(s.ctx)

	err = types.ValidateGenesis(exportedState)
	require.NoError(s.T(), err)

	require.ElementsMatch(s.T(), expectedState.OraclePrograms, exportedState.OraclePrograms)
	require.Equal(s.T(), expectedState.CoreContractRegistry, exportedState.CoreContractRegistry)
	require.Equal(s.T(), expectedState.Params, exportedState.Params)

	s.keeper.InitGenesis(s.ctx, exportedState)

	importedState := s.keeper.ExportGenesis(s.ctx)

	err = types.ValidateGenesis(importedState)
	require.NoError(s.T(), err)

	require.ElementsMatch(s.T(), exportedState.OraclePrograms, importedState.OraclePrograms)
	require.Equal(s.T(), exportedState.CoreContractRegistry, importedState.CoreContractRegistry)
	require.Equal(s.T(), exportedState.Params, importedState.Params)
}
