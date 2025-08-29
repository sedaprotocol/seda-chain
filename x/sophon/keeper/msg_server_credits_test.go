package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
