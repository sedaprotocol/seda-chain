package keeper_test

import (
	"encoding/hex"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"go.uber.org/mock/gomock"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) TestMsgServer_RegisterDataProxy() {
	fee := sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, types.DefaultRegistrationFee))
	adminAddr, err := sdk.AccAddressFromBech32("seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5")
	s.Require().NoError(err)

	tests := []struct {
		name      string
		msg       *types.MsgRegisterDataProxy
		expected  *types.ProxyConfig
		wantErr   error
		mockSetup func()
	}{
		{
			name: "Happy path",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1",
				Signature:     "6e8b21cf5fb2a87ea39d5320d37a47c3abdb70c41cafb0b6a499c0a7489ac04b22c9117bdd8f2a057818fd0baf5c1c529cb36f95037a28d20170dee66f322693",
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
				PubKey:        "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1",
				Signature:     "a8fe95113f86e564d63b0214aa44c4bdf793d0276ff516347e8d9a253a37aec9412ebb64c1d7e4d65e924c1da8b7d2b1f8afcf695b2148d31e31d3967c89b73e",
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
			expected:  nil,
			wantErr:   sdkerrors.ErrInvalidAddress,
			mockSetup: func() {},
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
			wantErr:  hex.InvalidByteError(byte('g')),
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
			wantErr:  hex.InvalidByteError(byte('g')),
		},
		{
			name: "Empty payout address",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4f3",
				Signature:     "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected:  nil,
			wantErr:   sdkerrors.ErrInvalidRequest,
			mockSetup: func() {},
		},
		{
			name: "Invalid fee denom",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "uatom",
					Amount: s.NewIntFromString("10000"),
				},
				Memo:      "",
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4f3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected:  nil,
			wantErr:   sdkerrors.ErrInvalidRequest,
			mockSetup: func() {},
		},
		{
			name: "Invalid memo length",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee: &sdk.Coin{
					Denom:  "uatom",
					Amount: s.NewIntFromString("10000"),
				},
				Memo:      strings.Repeat("a", types.MaxMemoLength+1),
				PubKey:    "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4f3",
				Signature: "5076d9d98754505d2f6f94f5a44062b9e95c2c5cfe7f21c69270814dc947bd285f5ed64e595aa956004687a225263f2831252cb41379cab2e3505b90f3da2701",
			},
			expected:  nil,
			wantErr:   sdkerrors.ErrInvalidRequest,
			mockSetup: func() {},
		},
		{
			name: "Insufficient balance for registration fee",
			msg: &types.MsgRegisterDataProxy{
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				PubKey:        "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1",
				Signature:     "6e8b21cf5fb2a87ea39d5320d37a47c3abdb70c41cafb0b6a499c0a7489ac04b22c9117bdd8f2a057818fd0baf5c1c529cb36f95037a28d20170dee66f322693",
			},
			expected: &types.ProxyConfig{
				PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				Fee:           s.NewFeeFromString("10000000000000000000"),
				Memo:          "",
				FeeUpdate:     nil,
				AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			},
			wantErr: sdkerrors.ErrInsufficientFunds,
			mockSetup: func() {
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), adminAddr, types.ModuleName, fee).
					Return(sdkerrors.ErrInsufficientFunds).
					MaxTimes(1)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockSetup != nil {
				tt.mockSetup()
			} else {
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), adminAddr, types.ModuleName, fee).Return(nil)
				s.bankKeeper.EXPECT().BurnCoins(gomock.Any(), types.ModuleName, fee).Return(nil)
			}

			res, err := s.msgSrvr.RegisterDataProxy(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}
			s.Require().NoError(err)

			pubKeyBytes, err := hex.DecodeString(tt.msg.PubKey)
			s.Require().NoError(err)

			proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
			s.Require().NoError(err)
			s.Require().Equal(tt.expected, &proxyConfig)
		})
	}

	s.Run("Registering an already existing data proxy should fail", func() {
		s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), adminAddr, types.ModuleName, fee).Return(nil).Times(2)
		s.bankKeeper.EXPECT().BurnCoins(gomock.Any(), types.ModuleName, fee).Return(nil).Times(2)

		msg := &types.MsgRegisterDataProxy{
			AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee:           s.NewFeeFromString("10000000000000000000"),
			Memo:          "",
			PubKey:        "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1",
			Signature:     "32473a31c221d8cf9ce4f6269369c9ce2b0ce9139f3dc93bd9e020bd76186c06087dc8d5367c0b6e7175a108694f2abc89b9675bea8ff35360fb8dc7225f8870",
		}

		_, err = s.msgSrvr.RegisterDataProxy(s.ctx, msg)
		s.Require().NoError(err)

		pubKeyBytes, err := hex.DecodeString(msg.PubKey)
		s.Require().NoError(err)

		proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
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
					NewFee: s.NewFeeFromString("1337"),
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
					NewFee:       s.NewFeeFromString("1337"),
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
			wantErr:  sdkerrors.ErrorInvalidSigner,
		},
		{
			name: "Update fee with invalid fee denom",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: types.DoNotModifyField,
				NewMemo:          types.DoNotModifyField,
				NewFee: &sdk.Coin{
					Denom:  "uatom",
					Amount: s.NewIntFromString("10000"),
				},
				PubKey: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: nil,
			wantErr:  sdkerrors.ErrInvalidRequest,
		},
		{
			name: "Update memo with invalid memo length",
			msg: &types.MsgEditDataProxy{
				Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				NewPayoutAddress: types.DoNotModifyField,
				NewMemo:          strings.Repeat("b", types.MaxMemoLength+1),
				NewFee: &sdk.Coin{
					Denom:  "uatom",
					Amount: s.NewIntFromString("10000"),
				},
				PubKey: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			},
			expected: nil,
			wantErr:  sdkerrors.ErrInvalidRequest,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			err = s.keeper.SetDataProxyConfig(s.ctx, pubKeyBytes, initialProxyConfig)
			s.Require().NoError(err)

			res, err := s.msgSrvr.EditDataProxy(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)

			pubKeyBytes, err := hex.DecodeString(tt.msg.PubKey)
			s.Require().NoError(err)

			proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
			s.Require().NoError(err)
			s.Require().Equal(tt.expected, &proxyConfig)

			if proxyConfig.FeeUpdate != nil {
				updateScheduled, err := s.keeper.HasFeeUpdate(s.ctx, proxyConfig.FeeUpdate.UpdateHeight, pubKeyBytes)
				s.Require().NoError(err)
				s.Require().True(updateScheduled)
			}
		})
	}

	s.Run("Updating the fee for a proxy that already has a pending update should cancel the old update", func() {
		firstUpdateHeight := types.DefaultMinFeeUpdateDelay + 100
		secondUpdateHeight := types.DefaultMinFeeUpdateDelay + 37

		err = s.keeper.SetDataProxyConfig(s.ctx, pubKeyBytes, initialProxyConfig)
		s.Require().NoError(err)

		firstMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1337"),
			FeeUpdateDelay:   firstUpdateHeight,
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		}

		firstRes, err := s.msgSrvr.EditDataProxy(s.ctx, firstMsg)
		s.Require().NoError(err)
		s.Require().NotNil(firstRes)

		firstProxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().NotNil(firstProxyConfig.FeeUpdate)

		firstUpdateScheduled, err := s.keeper.HasFeeUpdate(s.ctx, int64(firstUpdateHeight), pubKeyBytes)
		s.Require().NoError(err)
		s.Require().True(firstUpdateScheduled)

		secondMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1984"),
			FeeUpdateDelay:   secondUpdateHeight,
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		}

		secondRes, err := s.msgSrvr.EditDataProxy(s.ctx, secondMsg)
		s.Require().NoError(err)
		s.Require().NotNil(secondRes)

		secondProxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().NotNil(secondProxyConfig.FeeUpdate)

		secondUpdateScheduled, err := s.keeper.HasFeeUpdate(s.ctx, int64(secondUpdateHeight), pubKeyBytes)
		s.Require().NoError(err)
		s.Require().True(secondUpdateScheduled)

		firstUpdateNoLongerScheduled, err := s.keeper.HasFeeUpdate(s.ctx, int64(firstUpdateHeight), pubKeyBytes)
		s.Require().NoError(err)
		s.Require().False(firstUpdateNoLongerScheduled)
	})

	s.Run("Scheduling a fee update should fail if the max updates per block is reached", func() {
		// Mock bank keeper setup
		fee := sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, types.DefaultRegistrationFee))
		adminAddr, err := sdk.AccAddressFromBech32("seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5")
		s.Require().NoError(err)
		s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), adminAddr, types.ModuleName, fee).Return(nil)
		s.bankKeeper.EXPECT().BurnCoins(gomock.Any(), types.ModuleName, fee).Return(nil)

		err = s.keeper.SetDataProxyConfig(s.ctx, pubKeyBytes, initialProxyConfig)
		s.Require().NoError(err)

		updateHeight := int64(types.DefaultMinFeeUpdateDelay + uint32(s.ctx.BlockHeight()))

		// Set the max number of updates for the target block
		for i := 0; i < types.MaxUpdatesPerBlock-1; i++ {
			newPubkey := make([]byte, len(pubKeyBytes)+1)
			copy(newPubkey, pubKeyBytes)
			newPubkey[len(pubKeyBytes)] = byte(i)
			s.keeper.SetFeeUpdate(s.ctx, updateHeight, newPubkey)
		}
		// One manual so we can verify that after rescheduling a spot is available
		maxEdit, err := s.msgSrvr.EditDataProxy(s.ctx, &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1337"),
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		})
		s.Require().NoError(err)
		s.Require().NotNil(maxEdit)

		// Register a new proxy
		registerRes, err := s.msgSrvr.RegisterDataProxy(s.ctx, &types.MsgRegisterDataProxy{
			AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			PayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			Fee:           s.NewFeeFromString("10000000000000000000"),
			Memo:          "",
			PubKey:        "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1",
			Signature:     "6e8b21cf5fb2a87ea39d5320d37a47c3abdb70c41cafb0b6a499c0a7489ac04b22c9117bdd8f2a057818fd0baf5c1c529cb36f95037a28d20170dee66f322693",
		})
		s.Require().NoError(err)
		s.Require().NotNil(registerRes)

		// Try to schedule a fee update
		editMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1984"),
			PubKey:           "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1",
		}

		firstEditRes, err := s.msgSrvr.EditDataProxy(s.ctx, editMsg)
		s.Require().Nil(firstEditRes)
		s.Require().ErrorIs(err, types.ErrMaxUpdatesReached)

		// Reschedule the fee update of the manually added proxy
		maxReschedule, err := s.msgSrvr.EditDataProxy(s.ctx, &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			FeeUpdateDelay:   types.DefaultMinFeeUpdateDelay + 100,
			NewFee:           s.NewFeeFromString("100"),
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		})
		s.Require().NoError(err)
		s.Require().NotNil(maxReschedule)

		// Retry the update that failed before
		secondEditRes, err := s.msgSrvr.EditDataProxy(s.ctx, editMsg)
		s.Require().NoError(err)
		s.Require().NotNil(secondEditRes)
	})

	s.Run("Transferring admin address should allow the new address to submit changes", func() {
		err = s.keeper.SetDataProxyConfig(s.ctx, pubKeyBytes, initialProxyConfig)
		s.Require().NoError(err)

		editMsg := &types.MsgEditDataProxy{
			Sender:           "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			NewPayoutAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			NewMemo:          types.DoNotModifyField,
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		}
		failedEdit, err := s.msgSrvr.EditDataProxy(s.ctx, editMsg)
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
		s.Require().Nil(failedEdit)

		transferRes, err := s.msgSrvr.TransferAdmin(s.ctx, &types.MsgTransferAdmin{
			Sender:          "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewAdminAddress: "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
			PubKey:          "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		})
		s.Require().NoError(err)
		s.Require().NotNil(transferRes)

		successfulEdit, err := s.msgSrvr.EditDataProxy(s.ctx, editMsg)
		s.Require().NoError(err)
		s.Require().NotNil(successfulEdit)
	})
}

func (s *KeeperTestSuite) TestMsgServer_UpdateParamsErrors() {
	authority := s.keeper.GetAuthority()
	cases := []struct {
		name    string
		input   *types.MsgUpdateParams
		wantErr error
	}{
		{
			name: "invalid minimum update delay",
			input: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					MinFeeUpdateDelay: 0,
				},
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l",
				Params: types.Params{
					MinFeeUpdateDelay: 8000,
				},
			},
			wantErr: sdkerrors.ErrUnauthorized,
		},
	}

	for _, tt := range cases {
		s.Run(tt.name, func() {
			res, err := s.msgSrvr.UpdateParams(s.ctx, tt.input)
			s.Require().ErrorIs(err, tt.wantErr)
			s.Require().Nil(res)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_UpdateParams() {
	authority := s.keeper.GetAuthority()
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	initialProxyConfig := types.ProxyConfig{
		PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		Fee:           s.NewFeeFromString("9"),
		Memo:          "test",
		FeeUpdate:     nil,
		AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
	}

	s.Run("Updating the minimum delay should not affect existing updates", func() {
		// Register data proxy
		err = s.keeper.SetDataProxyConfig(s.ctx, pubKeyBytes, initialProxyConfig)
		s.Require().NoError(err)

		// Edit data proxy fee and verify it is scheduled with the default delay
		firstEditMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1337"),
			PubKey:           pubKeyHex,
		}

		firstEditRes, err := s.msgSrvr.EditDataProxy(s.ctx, firstEditMsg)
		s.Require().NoError(err)
		s.Require().NotNil(firstEditRes)

		firstProxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().NotNil(firstProxyConfig.FeeUpdate)

		firstUpdateScheduled, err := s.keeper.HasFeeUpdate(s.ctx, firstEditRes.FeeUpdateHeight, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().True(firstUpdateScheduled)

		// Update params, increasing the minimum delay
		_, err = s.msgSrvr.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: authority,
			Params: types.Params{
				MinFeeUpdateDelay: types.DefaultMinFeeUpdateDelay + 100,
			},
		})
		s.Require().NoError(err)

		// Verify update is still pending at the original height
		firstUpdateStillScheduled, err := s.keeper.HasFeeUpdate(s.ctx, firstEditRes.FeeUpdateHeight, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().True(firstUpdateStillScheduled)

		// Schedule a new fee update
		secondEditMsg := &types.MsgEditDataProxy{
			Sender:           "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			NewPayoutAddress: types.DoNotModifyField,
			NewMemo:          types.DoNotModifyField,
			NewFee:           s.NewFeeFromString("1984"),
			PubKey:           "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
		}

		secondEditRes, err := s.msgSrvr.EditDataProxy(s.ctx, secondEditMsg)
		s.Require().NoError(err)
		s.Require().NotNil(secondEditRes)

		secondProxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().NotNil(secondProxyConfig.FeeUpdate)

		// Verify the new update is scheduled and the old one is cancelled
		secondUpdateScheduled, err := s.keeper.HasFeeUpdate(s.ctx, secondEditRes.FeeUpdateHeight, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().True(secondUpdateScheduled)

		firstUpdateNoLongerScheduled, err := s.keeper.HasFeeUpdate(s.ctx, firstEditRes.FeeUpdateHeight, pubKeyBytes)
		s.Require().NoError(err)
		s.Require().False(firstUpdateNoLongerScheduled)
	})
}
