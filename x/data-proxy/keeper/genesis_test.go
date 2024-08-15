package keeper_test

import (
	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	genState := types.GenesisState{}

	s.keeper.InitGenesis(s.ctx, genState)
	// exportedGenState := s.keeper.ExportGenesis(s.ctx)
	// TODO
}
