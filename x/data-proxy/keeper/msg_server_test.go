package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

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
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
				Signature:     "628e5f1a2662872636c91fe2103602b2f0d5b0c3a52c5cc564171b424b902612048704f4a3349c70f0d0c618ecc65aa884c545e717d94be2272a4f2d6021fa6b",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				FeeUpdate:     nil,
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Happy path with memo",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("9000000000000000000"),
				Memo:          "This is a sweet proxy",
				PubKey:        "034c0f86f0cb61f9ddb47c4ba0b2ca0470962b5a1c50bee3a563184979672195f4",
				Signature:     "65b010f830dd52d54c940cec63140354e99484e4a2db9df3e0a7524a4bfaf87e146c82faddcba00df59e57dd774fb147994fbccea16be841e60e9791ccdbb4c4",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("9000000000000000000"),
				Memo:          "This is a sweet proxy",
				FeeUpdate:     nil,
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Invalid address",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				Signature:     "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidAddress,
		},
		{
			name: "Invalid signature",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("9000000000000000000"),
				Memo:          "",
				PubKey:        "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
				Signature:     "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidSignature,
		},
		{
			name: "Invalid pubkey hex",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4g3",
				Signature:     "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected: nil,
			wantErr:  types.ErrInvalidHex,
		},
		{
			name: "Invalid signature hex",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4f3",
				Signature:     "5076g9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
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

	s.Run("Registering an already existing data proxy should fail", func() {
		msg := &types.MsgRegisterDataProxy{
			AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee:           s.NewFeeFromString("10000000000000000000"),

			PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
		}

		_, err := s.msgSrvr.RegisterDataProxy(s.ctx, msg)
		s.Require().NoError(err)

		proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, msg.PubKey)
		s.Require().NoError(err)
		s.Require().Equal(&types.ProxyConfig{
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee:           s.NewFeeFromString("10000000000000000000"),
			Memo:          "",
			FeeUpdate:     nil,
			AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		}, &proxyConfig)

		res, err := s.msgSrvr.RegisterDataProxy(s.ctx, msg)
		s.Require().ErrorIs(err, types.ErrAlreadyExists)
		s.Require().Nil(res)
	})
}

func (s *KeeperTestSuite) TestMsgServer_EditDataProxy() {
	pubKeyBytes, err := hex.DecodeString("02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3")
	s.Require().NoError(err)

	initialProxyConfig := types.ProxyConfig{
		PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		Fee:           s.NewFeeFromString("9"),
		Memo:          "test",
		FeeUpdate:     nil,
		AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
	}

	tests := []struct {
		name     string
		msg      *types.MsgEditDataProxy
		expected *types.ProxyConfig
		wantErr  error
	}{
		{
			name: "Update payout address",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				NewMemo:          types.DoNotModifyField,
				PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("9"),
				Memo:          "test",
				FeeUpdate:     nil,
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Update memo",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: types.DoNotModifyField,
				NewMemo:          "",
				PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("9"),
				Memo:          "",
				FeeUpdate:     nil,
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Update fee",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: types.DoNotModifyField,
				NewMemo:          types.DoNotModifyField,
				NewFee:           s.NewFeeFromString("1337"),
				PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("9"),
				Memo:          "test",
				FeeUpdate: &types.FeeUpdate{
					NewFee: *s.NewFeeFromString("1337"),
					// Height in test is 0, so update height should be minimum
					UpdateHeight: int64(types.DefaultMinFeeUpdateDelay),
				},
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Update fee with valid custom delay",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: types.DoNotModifyField,
				NewMemo:          types.DoNotModifyField,
				NewFee:           s.NewFeeFromString("1337"),
				FeeUpdateDelay:   types.DefaultMinFeeUpdateDelay + 100,
				PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("9"),
				Memo:          "test",
				FeeUpdate: &types.FeeUpdate{
					NewFee:       *s.NewFeeFromString("1337"),
					UpdateHeight: int64(types.DefaultMinFeeUpdateDelay + 100),
				},
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: nil,
		},
		{
			name: "Update fee with invalid custom delay",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: types.DoNotModifyField,
				NewMemo:          types.DoNotModifyField,
				NewFee:           s.NewFeeFromString("1337"),
				FeeUpdateDelay:   1,
				PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: nil,
			wantErr:  types.ErrInvalidDelay,
		},
		{
			name: "Update from address that's not the admin",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				NewPayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				NewMemo:          types.DoNotModifyField,
				PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: nil,
			wantErr:  types.ErrUnauthorized,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			err = s.keeper.DataProxyConfigs.Set(s.ctx, pubKeyBytes, initialProxyConfig)
			s.Require().NoError(err)

			res, err := s.msgSrvr.EditDataProxy(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)

			proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, tt.msg.PubKey)
			s.Require().NoError(err)
			s.Require().Equal(tt.expected, &proxyConfig)

			if proxyConfig.FeeUpdate != nil {
				updateScheduled, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(proxyConfig.FeeUpdate.UpdateHeight, pubKeyBytes))
				s.Require().NoError(err)
				s.Require().True(updateScheduled)
			}
		})
	}

	s.Run("Updating the fee for a proxy that already has a pending update should cancel the old update", func() {
		s.SetupTest()

		firstUpdateHeight := int64(types.DefaultMinFeeUpdateDelay + 100)
		secondUpdateHeight := int64(types.DefaultMinFeeUpdateDelay + 37)

		err = s.keeper.DataProxyConfigs.Set(s.ctx, pubKeyBytes, initialProxyConfig)
		s.Require().NoError(err)

		firstMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1337"),
			FeeUpdateDelay:   uint32(firstUpdateHeight),
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		}

		firstRes, err := s.msgSrvr.EditDataProxy(s.ctx, firstMsg)
		s.Require().NoError(err)
		s.Require().NotNil(firstRes)

		firstProxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, firstMsg.PubKey)
		s.Require().NoError(err)
		s.Require().NotNil(firstProxyConfig.FeeUpdate)

		firstUpdateScheduled, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(firstUpdateHeight, pubKeyBytes))
		s.Require().NoError(err)
		s.Require().True(firstUpdateScheduled)

		secondMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1984"),
			FeeUpdateDelay:   uint32(secondUpdateHeight),
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		}

		secondRes, err := s.msgSrvr.EditDataProxy(s.ctx, secondMsg)
		s.Require().NoError(err)
		s.Require().NotNil(secondRes)

		secondProxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, secondMsg.PubKey)
		s.Require().NoError(err)
		s.Require().NotNil(secondProxyConfig.FeeUpdate)

		secondUpdateScheduled, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(secondUpdateHeight, pubKeyBytes))
		s.Require().NoError(err)
		s.Require().True(secondUpdateScheduled)

		firstUpdateNoLongerScheduled, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(firstUpdateHeight, pubKeyBytes))
		s.Require().NoError(err)
		s.Require().False(firstUpdateNoLongerScheduled)
	})
}
