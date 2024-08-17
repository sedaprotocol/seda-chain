package keeper_test

import (
	"encoding/hex"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) TestImportExportGenesis() {
	pubkeyOne, err := hex.DecodeString("034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4")
	s.Require().NoError(err)

	pubkeyTwo, err := hex.DecodeString("02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3")
	s.Require().NoError(err)

	genState := types.GenesisState{
		Params: types.DefaultParams(),
		DataProxyConfigs: []types.DataProxyConfig{
			{
				DataProxyPubkey: pubkeyOne,
				Config: &types.ProxyConfig{
					AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
					PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
					Fee:           s.NewFeeFromString("5"),
					Memo:          "",
					FeeUpdate:     nil,
				},
			},
			{
				DataProxyPubkey: pubkeyTwo,
				Config: &types.ProxyConfig{
					AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
					PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
					Fee:           s.NewFeeFromString("5000"),
					Memo:          "not my proxy friend",
					FeeUpdate: &types.FeeUpdate{
						NewFee:       *s.NewFeeFromString("10000"),
						UpdateHeight: 500,
					},
				},
			},
		},
		FeeUpdateQueue: []types.FeeUpdateQueueRecord{
			{
				DataProxyPubkey: pubkeyTwo,
				UpdateHeight:    500,
			},
		},
	}

	s.keeper.InitGenesis(s.ctx, genState)
	exportedGenState := s.keeper.ExportGenesis(s.ctx)
	s.Require().Equal(genState.Params, exportedGenState.Params)
	s.Require().ElementsMatch(genState.DataProxyConfigs, exportedGenState.DataProxyConfigs)
	s.Require().ElementsMatch(genState.FeeUpdateQueue, exportedGenState.FeeUpdateQueue)
}
