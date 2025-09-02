package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestMsgServer_SettleCreditsWithdraw() {
	adminAddr := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	adminAccAddr, err := sdk.AccAddressFromBech32(adminAddr)
	s.Require().NoError(err)

	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	initialCredits := math.NewInt(1000000000000000000)
	initialBalance := initialCredits.Mul(math.NewInt(2))

	tests := []struct {
		name               string
		msg                *types.MsgSettleCredits
		expected           *types.SophonInfo
		expectedBankAmount math.Int
		wantErr            error
	}{
		{
			name: "Happy path withdraw all credits",
			msg: &types.MsgSettleCredits{
				AdminAddress:    adminAddr,
				SophonPublicKey: pubKeyHex,
				Amount:          initialCredits,
				SettleType:      types.SETTLE_TYPE_WITHDRAW,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance.Sub(initialCredits),
				UsedCredits:  math.NewInt(0),
			},
			expectedBankAmount: initialCredits,
		},
		{
			name: "Happy path withdraw partial credits",
			msg: &types.MsgSettleCredits{
				AdminAddress:    adminAddr,
				SophonPublicKey: pubKeyHex,
				Amount:          initialCredits.Quo(math.NewInt(2)),
				SettleType:      types.SETTLE_TYPE_WITHDRAW,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance.Sub(initialCredits.Quo(math.NewInt(2))),
				UsedCredits:  initialCredits.Quo(math.NewInt(2)),
			},
			expectedBankAmount: initialCredits.Quo(math.NewInt(2)),
		},
		{
			name: "withdraw more than available credits",
			msg: &types.MsgSettleCredits{
				AdminAddress:    adminAddr,
				SophonPublicKey: pubKeyHex,
				Amount:          initialBalance.Add(math.NewInt(1)),
				SettleType:      types.SETTLE_TYPE_WITHDRAW,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance,
				UsedCredits:  initialBalance,
			},
			wantErr: types.ErrInsufficientCredits,
		},
		{
			name: "withdraw with invalid admin address",
			msg: &types.MsgSettleCredits{
				AdminAddress:    "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
				Amount:          initialCredits,
				SettleType:      types.SETTLE_TYPE_WITHDRAW,
			},
			wantErr: sdkerrors.ErrorInvalidSigner,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.keeper.SetSophonInfo(s.ctx, pubKey, types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance,
				UsedCredits:  initialCredits,
			})

			if test.wantErr == nil {
				s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, adminAccAddr, sdk.NewCoins(sdk.NewCoin("aseda", test.expectedBankAmount))).Return(nil)
			}

			response, err := s.msgSrvr.SettleCredits(s.ctx, test.msg)
			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				s.Require().Nil(response)
				return
			}

			s.Require().NoError(err)
			sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
			s.Require().NoError(err)

			s.Require().Equal(test.expected, &sophonInfo)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_SettleCreditsBurn() {
	adminAddr := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	initialCredits := math.NewInt(1000000000000000000)
	initialBalance := initialCredits.Mul(math.NewInt(2))

	tests := []struct {
		name               string
		msg                *types.MsgSettleCredits
		expected           *types.SophonInfo
		expectedBankAmount math.Int
		wantErr            error
	}{
		{
			name: "Happy path burn all credits",
			msg: &types.MsgSettleCredits{
				AdminAddress:    adminAddr,
				SophonPublicKey: pubKeyHex,
				Amount:          initialCredits,
				SettleType:      types.SETTLE_TYPE_BURN,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance.Sub(initialCredits),
				UsedCredits:  math.NewInt(0),
			},
			expectedBankAmount: initialCredits,
		},
		{
			name: "Happy path burn partial credits",
			msg: &types.MsgSettleCredits{
				AdminAddress:    adminAddr,
				SophonPublicKey: pubKeyHex,
				Amount:          initialCredits.Quo(math.NewInt(2)),
				SettleType:      types.SETTLE_TYPE_BURN,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance.Sub(initialCredits.Quo(math.NewInt(2))),
				UsedCredits:  initialCredits.Quo(math.NewInt(2)),
			},
			expectedBankAmount: initialCredits.Quo(math.NewInt(2)),
		},
		{
			name: "burn more than available credits",
			msg: &types.MsgSettleCredits{
				AdminAddress:    adminAddr,
				SophonPublicKey: pubKeyHex,
				Amount:          initialBalance.Add(math.NewInt(1)),
				SettleType:      types.SETTLE_TYPE_BURN,
			},
			expected: &types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance,
				UsedCredits:  initialBalance,
			},
			wantErr: types.ErrInsufficientCredits,
		},
		{
			name: "burn with invalid admin address",
			msg: &types.MsgSettleCredits{
				AdminAddress:    "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh",
				SophonPublicKey: pubKeyHex,
				Amount:          initialCredits,
				SettleType:      types.SETTLE_TYPE_BURN,
			},
			wantErr: sdkerrors.ErrorInvalidSigner,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.keeper.SetSophonInfo(s.ctx, pubKey, types.SophonInfo{
				Id:           0,
				OwnerAddress: adminAddr,
				AdminAddress: adminAddr,
				Address:      adminAddr,
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      initialBalance,
				UsedCredits:  initialCredits,
			})

			if test.wantErr == nil {
				s.bankKeeper.EXPECT().BurnCoins(s.ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("aseda", test.expectedBankAmount))).Return(nil)
			}

			response, err := s.msgSrvr.SettleCredits(s.ctx, test.msg)
			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				s.Require().Nil(response)
				return
			}

			s.Require().NoError(err)
			sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
			s.Require().NoError(err)

			s.Require().Equal(test.expected, &sophonInfo)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_SubmitReports() {
	adminAddr := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"

	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	user1 := "user1"
	user1Credits := math.NewInt(1000000000000000000)
	user2 := "user2"
	user2Credits := math.NewInt(8000000000000000000)

	initialBalance := user1Credits.Add(user2Credits)

	dataProxy1 := "021dd035f760061e2833581d4ab50440a355db0ac98e489bf63a5dbc0e89e4af79"
	dataProxy1PubKey, err := hex.DecodeString(dataProxy1)
	s.Require().NoError(err)
	dataProxy1PayoutAddress := "seda1wyzxdtpl0c99c92n397r3drlhj09qfjvf6teyh"
	dataProxy1Price := math.NewInt(100000)
	dataProxy1Price2 := math.NewInt(200)
	dataProxy2 := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	dataProxy2PubKey, err := hex.DecodeString(dataProxy2)
	s.Require().NoError(err)
	dataProxy2PayoutAddress := "seda1xd04svzj6zj93g4eknhp6aq2yyptagcc2zeetj"
	dataProxy2Price := math.NewInt(1)

	setupTest := func() {
		s.keeper.SetSophonInfo(s.ctx, pubKey, types.SophonInfo{
			Id:           0,
			OwnerAddress: adminAddr,
			AdminAddress: adminAddr,
			Address:      adminAddr,
			PublicKey:    pubKey,
			Memo:         "",
			Balance:      initialBalance,
			UsedCredits:  math.NewInt(0),
		})

		s.keeper.SetSophonUser(s.ctx, 0, user1, types.SophonUser{
			UserId:  user1,
			Credits: user1Credits,
		})

		s.keeper.SetSophonUser(s.ctx, 0, user2, types.SophonUser{
			UserId:  user2,
			Credits: user2Credits,
		})

		s.dataProxyKeeper.EXPECT().GetDataProxyConfig(s.ctx, dataProxy1PubKey).Return(dataproxytypes.ProxyConfig{
			PayoutAddress: dataProxy1PayoutAddress,
		}, nil).AnyTimes()

		s.dataProxyKeeper.EXPECT().GetDataProxyConfig(s.ctx, dataProxy2PubKey).Return(dataproxytypes.ProxyConfig{
			PayoutAddress: dataProxy2PayoutAddress,
		}, nil).AnyTimes()
	}

	s.Run("Single user all credits", func() {
		setupTest()

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: user1Credits,
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		s.Require().Equal(sophonInfo.Balance, initialBalance)
		s.Require().Equal(sophonInfo.UsedCredits, user1Credits)

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user1)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, math.NewInt(0))
	})

	s.Run("Single user partial credits", func() {
		setupTest()

		reportedCredits := math.NewInt(6)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCredits,
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		s.Require().Equal(sophonInfo.Balance, initialBalance)
		s.Require().Equal(sophonInfo.UsedCredits, reportedCredits)

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user2)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user2Credits.Sub(reportedCredits))
	})

	s.Run("Single user, single data proxy report", func() {
		setupTest()

		reportedCredits := math.NewInt(90000000)
		dataProxy1Amount := uint64(5)

		expectedDataProxy1Credits := dataProxy1Price.Mul(math.NewIntFromUint64(dataProxy1Amount))
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy1PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy1Credits))).Return(nil)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCredits,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		s.Require().Equal(sophonInfo.Balance, initialBalance.Sub(expectedDataProxy1Credits))
		s.Require().Equal(sophonInfo.UsedCredits, reportedCredits.Sub(expectedDataProxy1Credits))

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user1)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user1Credits.Sub(reportedCredits))
	})

	s.Run("Single user, multiple data proxy reports", func() {
		setupTest()

		reportedCredits := math.NewInt(90000000)
		dataProxy1Amount := uint64(5)
		dataProxy2Amount := uint64(10)

		expectedDataProxy1Credits := dataProxy1Price.Mul(math.NewIntFromUint64(dataProxy1Amount))
		expectedDataProxy2Credits := dataProxy2Price.Mul(math.NewIntFromUint64(dataProxy2Amount))
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy1PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy1Credits))).Return(nil)
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy2PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy2Credits))).Return(nil)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCredits,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
						{
							DataProxyPubKey: dataProxy2,
							Amount:          dataProxy2Amount,
							Price:           dataProxy2Price,
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		s.Require().Equal(sophonInfo.Balance, initialBalance.Sub(expectedDataProxy1Credits).Sub(expectedDataProxy2Credits))
		s.Require().Equal(sophonInfo.UsedCredits, reportedCredits.Sub(expectedDataProxy1Credits).Sub(expectedDataProxy2Credits))

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user2)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user2Credits.Sub(reportedCredits))
	})

	// Same data proxy, different prices
	s.Run("Single user, same data proxy, different prices", func() {
		setupTest()

		reportedCredits := math.NewInt(90000000)
		dataProxy1Amount := uint64(5)
		dataProxy1Amount2 := uint64(10)

		expectedDataProxy1Credits := dataProxy1Price.Mul(math.NewIntFromUint64(dataProxy1Amount))
		expectedDataProxy1Credits2 := dataProxy1Price2.Mul(math.NewIntFromUint64(dataProxy1Amount2))
		totalExpectedDataProxy1Credits := expectedDataProxy1Credits.Add(expectedDataProxy1Credits2)
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx,
			types.ModuleName,
			sdk.MustAccAddressFromBech32(dataProxy1PayoutAddress),
			sdk.NewCoins(sdk.NewCoin("aseda", totalExpectedDataProxy1Credits)),
		).Return(nil)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCredits,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount2,
							Price:           dataProxy1Price2,
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		s.Require().Equal(sophonInfo.Balance, initialBalance.Sub(totalExpectedDataProxy1Credits))
		s.Require().Equal(sophonInfo.UsedCredits, reportedCredits.Sub(totalExpectedDataProxy1Credits))

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user2)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user2Credits.Sub(reportedCredits))
	})

	s.Run("Multiple users, multiple data proxy reports, same proxy shared between users", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)
		reportedCreditsUser2 := math.NewInt(87890000000)
		dataProxy1Amount := uint64(5)
		dataProxy2Amount := uint64(10)

		// Shared between users
		expectedDataProxy1Credits := dataProxy1Price.Mul(math.NewIntFromUint64(dataProxy1Amount)).Mul(math.NewInt(2))
		expectedDataProxy2Credits := dataProxy2Price.Mul(math.NewIntFromUint64(dataProxy2Amount))
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy1PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy1Credits))).Return(nil)
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy2PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy2Credits))).Return(nil)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
					},
				},
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCreditsUser2,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
						{
							DataProxyPubKey: dataProxy2,
							Amount:          dataProxy2Amount,
							Price:           dataProxy2Price,
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		s.Require().Equal(sophonInfo.Balance, initialBalance.Sub(expectedDataProxy1Credits).Sub(expectedDataProxy2Credits))
		s.Require().Equal(sophonInfo.UsedCredits, reportedCreditsUser1.Add(reportedCreditsUser2).Sub(expectedDataProxy1Credits).Sub(expectedDataProxy2Credits))

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user1)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user1Credits.Sub(reportedCreditsUser1))

		sophonUser, err = s.keeper.GetSophonUser(s.ctx, 0, user2)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user2Credits.Sub(reportedCreditsUser2))
	})

	s.Run("Insufficient credits", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)
		reportedCreditsUser2 := math.NewInt(9000000000000000000)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          1,
							Price:           dataProxy1Price,
						},
					},
				},
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCreditsUser2,
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().ErrorIs(err, types.ErrInsufficientCredits)
		s.Require().Nil(response)
	})

	s.Run("Non-existent sophon", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          1,
							Price:           dataProxy1Price,
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
		s.Require().Nil(response)
	})

	s.Run("Non-existent user", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      "non-existent-user",
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          1,
							Price:           dataProxy1Price,
						},
					},
				},
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
		s.Require().Nil(response)
	})

	s.Run("Multiple users, multiple data proxy reports, 1 non-existent data proxy", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)
		reportedCreditsUser2 := math.NewInt(87890000000)
		dataProxy1Amount := uint64(5)
		unknownDataProxyPubKey := "01"
		unknownDataProxyPubKeyBytes, err := hex.DecodeString(unknownDataProxyPubKey)
		s.Require().NoError(err)

		expectedDataProxy1Credits := dataProxy1Price.Mul(math.NewIntFromUint64(dataProxy1Amount))
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy1PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy1Credits))).Return(nil)
		s.dataProxyKeeper.EXPECT().GetDataProxyConfig(s.ctx, unknownDataProxyPubKeyBytes).Return(dataproxytypes.ProxyConfig{}, collections.ErrNotFound)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
					},
				},
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCreditsUser2,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: unknownDataProxyPubKey,
							Amount:          5,
							Price:           math.NewInt(999),
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		// Only the known data proxy is subtracted from the balance.
		s.Require().Equal(sophonInfo.Balance, initialBalance.Sub(expectedDataProxy1Credits))
		s.Require().Equal(sophonInfo.UsedCredits, reportedCreditsUser1.Add(reportedCreditsUser2).Sub(expectedDataProxy1Credits))

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user1)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user1Credits.Sub(reportedCreditsUser1))

		sophonUser, err = s.keeper.GetSophonUser(s.ctx, 0, user2)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user2Credits.Sub(reportedCreditsUser2))
	})

	s.Run("Multiple users, multiple data proxy reports, 1 invalid data proxy payout address", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)
		reportedCreditsUser2 := math.NewInt(87890000000)
		dataProxy1Amount := uint64(5)
		invalidDataProxyPubKey := "01"
		invalidDataProxyPubKeyBytes, err := hex.DecodeString(invalidDataProxyPubKey)
		s.Require().NoError(err)

		expectedDataProxy1Credits := dataProxy1Price.Mul(math.NewIntFromUint64(dataProxy1Amount))
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(s.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(dataProxy1PayoutAddress), sdk.NewCoins(sdk.NewCoin("aseda", expectedDataProxy1Credits))).Return(nil)
		s.dataProxyKeeper.EXPECT().GetDataProxyConfig(s.ctx, invalidDataProxyPubKeyBytes).Return(dataproxytypes.ProxyConfig{
			PayoutAddress: "invalid-address",
		}, nil)

		msg := &types.MsgSubmitReports{
			Address:         adminAddr,
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          dataProxy1Amount,
							Price:           dataProxy1Price,
						},
					},
				},
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCreditsUser2,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: invalidDataProxyPubKey,
							Amount:          5,
							Price:           math.NewInt(999),
						},
					},
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().NoError(err)
		s.Require().Equal(response, &types.MsgSubmitReportsResponse{})

		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
		s.Require().NoError(err)
		// Only the data proxy with a valid payout address is subtracted from the balance.
		s.Require().Equal(sophonInfo.Balance, initialBalance.Sub(expectedDataProxy1Credits))
		s.Require().Equal(sophonInfo.UsedCredits, reportedCreditsUser1.Add(reportedCreditsUser2).Sub(expectedDataProxy1Credits))

		sophonUser, err := s.keeper.GetSophonUser(s.ctx, 0, user1)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user1Credits.Sub(reportedCreditsUser1))

		sophonUser, err = s.keeper.GetSophonUser(s.ctx, 0, user2)
		s.Require().NoError(err)
		s.Require().Equal(sophonUser.Credits, user2Credits.Sub(reportedCreditsUser2))
	})

	s.Run("Wrong signer address", func() {
		setupTest()

		reportedCreditsUser1 := math.NewInt(90000000)

		msg := &types.MsgSubmitReports{
			Address:         "seda10d07y265gmmuvt4z0w9aw880jnsr700jvvla4j",
			SophonPublicKey: pubKey,
			Reports: []*types.UserReport{
				{
					UserId:      user1,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
					DataProxyReports: []*types.DataProxyReport{
						{
							DataProxyPubKey: dataProxy1,
							Amount:          1,
							Price:           dataProxy1Price,
						},
					},
				},
				{
					UserId:      user2,
					Queries:     100,
					UsedCredits: reportedCreditsUser1,
				},
			},
		}

		response, err := s.msgSrvr.SubmitReports(s.ctx, msg)
		s.Require().ErrorIs(err, sdkerrors.ErrorInvalidSigner)
		s.Require().Nil(response)
	})
}
