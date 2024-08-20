package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) TestQuerier_ProxyConfig() {
	tests := []struct {
		name      string
		config    *types.ProxyConfig
		pubKeyHex string
		wantErr   error
	}{
		{
			name: "Simple proxy",
			config: &types.ProxyConfig{
				AdminAddress:  "admin",
				PayoutAddress: "pay",
				Fee:           &sdk.Coin{Denom: "aseda", Amount: math.NewInt(5)},
				Memo:          "",
				FeeUpdate:     nil,
			},
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   nil,
		},
		{
			name: "Proxy with pending update",
			config: &types.ProxyConfig{
				AdminAddress:  "admin",
				PayoutAddress: "pay",
				Fee:           &sdk.Coin{Denom: "aseda", Amount: math.NewInt(5)},
				Memo:          "",
				FeeUpdate: &types.FeeUpdate{
					NewFee:       sdk.Coin{Denom: "aseda", Amount: math.NewInt(10)},
					UpdateHeight: 6,
				},
			},
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   nil,
		},
		{
			name:      "Unknown pubkey",
			config:    nil,
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   sdkerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			if tt.config != nil {
				pubkeyBytes, err := hex.DecodeString(tt.pubKeyHex)
				s.Require().NoError(err)

				err = s.keeper.DataProxyConfigs.Set(s.ctx, pubkeyBytes, *tt.config)
				s.Require().NoError(err)
			}

			res, err := s.queryClient.DataProxyConfig(s.ctx, &types.QueryDataProxyConfigRequest{PubKey: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"})
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(tt.config, res.Config)
		})
	}
}
