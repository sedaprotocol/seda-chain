package keeper_test

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sedaprotocol/seda-chain/x/sophon/types"
)

func (s *KeeperTestSuite) TestQuerier_SophonInfo() {
	pubKeyHex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	tests := []struct {
		name       string
		sophonInfo *types.SophonInfo
		pubKeyHex  string
		wantErr    error
	}{
		{
			name: "Existing sophon",
			sophonInfo: &types.SophonInfo{
				Id:           0,
				OwnerAddress: "owner",
				AdminAddress: "admin",
				Address:      "address",
				PublicKey:    pubKey,
				Memo:         "memo",
				Balance:      math.NewInt(0),
				UsedCredits:  math.NewInt(0),
			},
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   nil,
		},
		{
			name:      "Unknown sophon",
			pubKeyHex: "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3",
			wantErr:   sdkerrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.sophonInfo != nil {
				pubkeyBytes, err := hex.DecodeString(tt.pubKeyHex)
				s.Require().NoError(err)

				err = s.keeper.SetSophonInfo(s.ctx, pubkeyBytes, *tt.sophonInfo)
				s.Require().NoError(err)
			}

			res, err := s.queryClient.SophonInfo(s.ctx, &types.QuerySophonInfoRequest{SophonPubKey: tt.pubKeyHex})
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(*tt.sophonInfo, res.Info)
		})
	}
}

func (s *KeeperTestSuite) TestQuerier_SophonTransfer() {
	pubKeyHex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	newOwnerAddress := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	newOwnerAddressBz, err := sdk.AccAddressFromBech32(newOwnerAddress)
	s.Require().NoError(err)

	s.Run("No registered sophon", func() {
		res, err := s.queryClient.SophonTransfer(s.ctx, &types.QuerySophonTransferRequest{SophonPubKey: pubKeyHex})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
		s.Require().Nil(res)
	})

	s.Run("Pending sophon transfer", func() {
		sophonInfo, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
			OwnerAddress: "owner",
			AdminAddress: "admin",
			Address:      "address",
			PublicKey:    pubKey,
			Memo:         "memo",
		})
		s.Require().NoError(err)

		err = s.keeper.SetSophonTransfer(s.ctx, sophonInfo.Id, newOwnerAddressBz)
		s.Require().NoError(err)
		res, err := s.queryClient.SophonTransfer(s.ctx, &types.QuerySophonTransferRequest{SophonPubKey: pubKeyHex})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().Equal(newOwnerAddress, res.NewOwnerAddress)
	})

	s.Run("No pending transfer", func() {
		_, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
			OwnerAddress: "owner",
			AdminAddress: "admin",
			Address:      "address",
			PublicKey:    pubKey,
			Memo:         "memo",
		})
		s.Require().NoError(err)

		res, err := s.queryClient.SophonTransfer(s.ctx, &types.QuerySophonTransferRequest{SophonPubKey: pubKeyHex})
		s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
		s.Require().Nil(res)
	})
}

func (s *KeeperTestSuite) TestQuerier_SophonUsers() {
	// 10 users
	pubKey1Hex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey1, err := hex.DecodeString(pubKey1Hex)
	s.Require().NoError(err)

	// 5 users
	pubKey2Hex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	pubKey2, err := hex.DecodeString(pubKey2Hex)
	s.Require().NoError(err)

	// No users
	pubKey3Hex := "021e9706a29a6c22af65d7051484d02e093543e853e1ab75362a4666e8a14ee7b4"
	pubKey3, err := hex.DecodeString(pubKey3Hex)
	s.Require().NoError(err)

	makeSophonUser := func(sophonID uint64, index int, creditsMultiplier int64) types.SophonUser {
		return types.SophonUser{
			UserId:  fmt.Sprintf("sophon_%d_user_%d", sophonID, index),
			Credits: math.NewInt(int64(index) * creditsMultiplier),
		}
	}

	testSetup := func() {
		sophonInfos := make([]types.SophonInfo, 0)
		for i, pubKey := range [][]byte{pubKey1, pubKey2, pubKey3} {
			sophonInfo, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
				OwnerAddress: fmt.Sprintf("owner_%d", i),
				AdminAddress: fmt.Sprintf("admin_%d", i),
				Address:      fmt.Sprintf("address_%d", i),
				PublicKey:    pubKey,
				Memo:         fmt.Sprintf("memo_%d", i),
			})
			s.Require().NoError(err)
			sophonInfos = append(sophonInfos, sophonInfo)
		}

		for i := range 10 {
			sophonUser := makeSophonUser(0, i, 33)
			err := s.keeper.SetSophonUser(s.ctx, sophonInfos[0].Id, sophonUser.UserId, sophonUser)
			s.Require().NoError(err)
		}

		for i := range 5 {
			sophonUser := makeSophonUser(1, i, 100)
			err := s.keeper.SetSophonUser(s.ctx, sophonInfos[1].Id, sophonUser.UserId, sophonUser)
			s.Require().NoError(err)
		}
	}

	tests := []struct {
		name          string
		pubKeyHex     string
		wantErr       error
		pagination    *query.PageRequest
		wantUsers     []types.SophonUser
		wantUsersNext []types.SophonUser
	}{
		{
			name:      "Unknown sophon",
			pubKeyHex: "03cbede4de965ab2ac6b9468b97d0045fc092dbea6b88407301062c531dac644b8",
			wantErr:   sdkerrors.ErrNotFound,
		},
		{
			name:      "Sophon with no users",
			pubKeyHex: pubKey3Hex,
			wantUsers: nil,
		},
		{
			name:      "Sophon with 5 users, no pagination",
			pubKeyHex: pubKey2Hex,
			wantUsers: func() []types.SophonUser {
				users := make([]types.SophonUser, 0)
				for i := range 5 {
					users = append(users, makeSophonUser(1, i, 100))
				}
				return users
			}(),
		},
		{
			name:      "Sophon with 10 users, pagination",
			pubKeyHex: pubKey1Hex,
			pagination: &query.PageRequest{
				Key:        []byte{},
				Offset:     0,
				Limit:      5,
				CountTotal: false,
			},
			wantUsers: func() []types.SophonUser {
				users := make([]types.SophonUser, 0)
				for i := range 5 {
					users = append(users, makeSophonUser(0, i, 33))
				}
				return users
			}(),
			wantUsersNext: func() []types.SophonUser {
				users := make([]types.SophonUser, 0)
				for i := range 5 {
					users = append(users, makeSophonUser(0, i+5, 33))
				}
				return users
			}(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			testSetup()

			res, err := s.queryClient.SophonUsers(s.ctx, &types.QuerySophonUsersRequest{
				SophonPubKey: tt.pubKeyHex,
				Pagination:   tt.pagination,
			})
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().NotNil(res.Pagination)
			s.Require().Equal(tt.wantUsers, res.Users)

			if tt.wantUsersNext != nil {
				s.T().Logf("next key: %d", res.Pagination.Total)
				res, err := s.queryClient.SophonUsers(s.ctx, &types.QuerySophonUsersRequest{
					SophonPubKey: tt.pubKeyHex,
					Pagination: &query.PageRequest{
						Key:   res.Pagination.NextKey,
						Limit: tt.pagination.Limit,
					},
				})
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().NotNil(res.Pagination)
				s.Require().Equal(tt.wantUsersNext, res.Users)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQuerier_SophonUser() {
	// 5 users
	pubKey1Hex := "02095af5db08cef43871a4aa48a80bdddc5249e4234e7432c3d7eca14f31261b10"
	pubKey1, err := hex.DecodeString(pubKey1Hex)
	s.Require().NoError(err)

	// No users
	pubKey2Hex := "02100efce2a783cc7a3fbf9c5d15d4cc6e263337651312f21a35d30c16cb38f4c3"
	pubKey2, err := hex.DecodeString(pubKey2Hex)
	s.Require().NoError(err)

	makeSophonUser := func(index int, creditsMultiplier int64) types.SophonUser {
		return types.SophonUser{
			UserId:  fmt.Sprintf("user_%d", index),
			Credits: math.NewInt(int64(index) * creditsMultiplier),
		}
	}

	testSetup := func() {
		sophonInfos := make([]types.SophonInfo, 0)
		for i, pubKey := range [][]byte{pubKey1, pubKey2} {
			sophonInfo, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
				OwnerAddress: fmt.Sprintf("owner_%d", i),
				AdminAddress: fmt.Sprintf("admin_%d", i),
				Address:      fmt.Sprintf("address_%d", i),
				PublicKey:    pubKey,
				Memo:         fmt.Sprintf("memo_%d", i),
			})
			s.Require().NoError(err)
			sophonInfos = append(sophonInfos, sophonInfo)
		}

		for i := range 5 {
			err := s.keeper.SetSophonUser(s.ctx, sophonInfos[0].Id, fmt.Sprintf("user_%d", i), makeSophonUser(i, 100))
			s.Require().NoError(err)
		}
	}

	tests := []struct {
		name      string
		pubKeyHex string
		wantErr   error
		userId    string
		wantUser  types.SophonUser
	}{
		{
			name:      "Unknown sophon",
			pubKeyHex: "03cbede4de965ab2ac6b9468b97d0045fc092dbea6b88407301062c531dac644b8",
			userId:    "user_0",
			wantErr:   sdkerrors.ErrNotFound,
		},
		{
			name:      "Sophon with no users, user exists in different sophon",
			pubKeyHex: pubKey2Hex,
			userId:    "user_0",
			wantErr:   sdkerrors.ErrNotFound,
		},
		{
			name:      "Sophon with 5 users, user does not exist",
			pubKeyHex: pubKey1Hex,
			userId:    "user_10",
			wantErr:   sdkerrors.ErrNotFound,
		},
		{
			name:      "Sophon with 5 users, user exists",
			pubKeyHex: pubKey1Hex,
			userId:    "user_1",
			wantUser:  makeSophonUser(1, 100),
		},
		{
			name:      "Sophon with 5 users, different user exists",
			pubKeyHex: pubKey1Hex,
			userId:    "user_4",
			wantUser:  makeSophonUser(4, 100),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			testSetup()

			res, err := s.queryClient.SophonUser(s.ctx, &types.QuerySophonUserRequest{
				SophonPubKey: tt.pubKeyHex,
				UserId:       tt.userId,
			})
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(tt.wantUser, res.User)
		})
	}
}
