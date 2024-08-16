package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) NewIntFromString(val string) math.Int {
	amount, success := math.NewIntFromString(val)
	s.Require().True(success)
	return amount
}

func (s *KeeperTestSuite) TestMsgServer_RegisterDataProxy() {
	tests := []struct {
		name     string
		msg      *types.MsgRegisterDataProxy
		expected *types.ProxyConfig
		wantErr  error
	}{
		{
			name: "Happy path",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:      "",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:         "",
				FeeUpdate:    nil,
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Happy path with memo",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:      "This is a sweet proxy",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:         "This is a sweet proxy",
				FeeUpdate:    nil,
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Invalid address",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:      "",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidAddress,
		},
		{
			name: "Invalid signature",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("9000000000000000000"),
				},
				Memo:      "",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidSignature,
		},
		{
			name: "Invalid pubkey hex",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:      "",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidHex,
		},
		{
			name: "Invalid signature hex",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "aseda",
					Amount: s.NewIntFromString("10000000000000000000"),
				},
				Memo:      "",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4f3",
				Signature: "5076g9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidHex,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			res, err := s.msgSrvr.RegisterDataProxy(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)

			proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, tt.msg.PubKey)
			s.Require().NoError(err)
			s.Require().Equal(tt.expected, &proxyConfig)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_RegisterDataProxyDuplicate() {
	s.Run("Registering an already existing data proxy should fail", func() {
		msg := &types.MsgRegisterDataProxy{
			AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee: &sdk.Coin{
				Denom:  "aseda",
				Amount: s.NewIntFromString("10000000000000000000"),
			},
			Memo:      "",
			PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
		}

		_, err := s.msgSrvr.RegisterDataProxy(s.ctx, msg)
		s.Require().NoError(err)

		proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, msg.PubKey)
		s.Require().NoError(err)
		s.Require().Equal(&types.ProxyConfig{
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee: &sdk.Coin{
				Denom:  "aseda",
				Amount: s.NewIntFromString("10000000000000000000"),
			},
			Memo:         "",
			FeeUpdate:    nil,
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		}, &proxyConfig)

		res, err := s.msgSrvr.RegisterDataProxy(s.ctx, msg)
		s.Require().ErrorIs(err, types.ErrAlreadyExists)
		s.Require().Nil(res)
	})
}
