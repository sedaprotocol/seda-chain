package keeper_test

import (
	"encoding/hex"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"go.uber.org/mock/gomock"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestMsgServer_AddUser() {
	adminAddress := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	tests := []struct {
		name         string
		msg          *types.MsgAddUser
		expectedInfo *types.SophonInfo
		expectedUser *types.SophonUser
		wantErr      error
		mockSetup    func()
	}{
		{
			name: "Happy path no initial credits",
			msg: &types.MsgAddUser{
				AdminAddress:    adminAddress,
				SophonPublicKey: pubKeyHex,
				UserId:          "user1",
				InitialCredits:  math.NewInt(0),
			},
			expectedInfo: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      math.NewInt(0),
				UsedCredits:  math.NewInt(0),
			},
			expectedUser: &types.SophonUser{
				UserId:  "user1",
				Credits: math.NewInt(0),
			},
		},
		{
			name: "Happy path with initial credits",
			msg: &types.MsgAddUser{
				AdminAddress:    adminAddress,
				SophonPublicKey: pubKeyHex,
				UserId:          "user1",
				InitialCredits:  math.NewInt(10_000_000),
			},
			expectedInfo: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKey,
				Memo:         "",
				Balance:      math.NewInt(10_000_000),
				UsedCredits:  math.NewInt(0),
			},
			expectedUser: &types.SophonUser{
				UserId:  "user1",
				Credits: math.NewInt(10_000_000),
			},
			mockSetup: func() {
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "Not enough balance for initial credits",
			msg: &types.MsgAddUser{
				AdminAddress:    adminAddress,
				SophonPublicKey: pubKeyHex,
				UserId:          "user1",
				InitialCredits:  math.NewInt(10_000_000),
			},
			wantErr: sdkerrors.ErrInsufficientFunds,
			mockSetup: func() {
				adminAddr, err := sdk.AccAddressFromBech32(adminAddress)
				s.Require().NoError(err)

				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), adminAddr, types.ModuleName, sdk.NewCoins(sdk.NewCoin("aseda", math.NewInt(10_000_000)))).
					Return(sdkerrors.ErrInsufficientFunds).
					MaxTimes(1)
			},
		},
		{
			name: "Not the expected admin address",
			msg: &types.MsgAddUser{
				AdminAddress:    "seda1jd2q0mz0vzs75tp7lyuzf9064zccddgs8utjr5",
				SophonPublicKey: pubKeyHex,
				UserId:          "user1",
				InitialCredits:  math.NewInt(10_000_000),
			},
			wantErr: sdkerrors.ErrorInvalidSigner,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			if test.mockSetup != nil {
				test.mockSetup()
			}

			_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
				Authority:    s.keeper.GetAuthority(),
				OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				PublicKey:    pubKeyHex,
				Memo:         "",
			})
			s.Require().NoError(err)

			response, err := s.msgSrvr.AddUser(s.ctx, test.msg)
			if test.wantErr != nil {
				s.Require().ErrorIs(err, test.wantErr)
				s.Require().Nil(response)
				return
			}
			s.Require().NoError(err)

			sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, pubKey)
			s.Require().NoError(err)
			s.Require().Equal(test.expectedInfo, &sophonInfo)

			sophonUser, err := s.keeper.GetSophonUser(s.ctx, sophonInfo.Id, test.msg.UserId)
			s.Require().NoError(err)
			s.Require().Equal(test.expectedUser, &sophonUser)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_AddUser_AlreadyExists() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"

	_, err := s.msgSrvr.RegisterSophon(s.ctx, &types.MsgRegisterSophon{
		Authority:    s.keeper.GetAuthority(),
		OwnerAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		Address:      "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		PublicKey:    pubKeyHex,
		Memo:         "",
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.AddUser(s.ctx, &types.MsgAddUser{
		AdminAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		SophonPublicKey: pubKeyHex,
		UserId:          "user1",
		InitialCredits:  math.NewInt(0),
	})
	s.Require().NoError(err)

	_, err = s.msgSrvr.AddUser(s.ctx, &types.MsgAddUser{
		AdminAddress:    "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		SophonPublicKey: pubKeyHex,
		UserId:          "user1",
		InitialCredits:  math.NewInt(10000),
	})
	s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)
}
