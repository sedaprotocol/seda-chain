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

func (s *KeeperTestSuite) TestQuerier_SophonEligibility() {
	adminAddress := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"

	pubKeyHex := "041b84c5567b126440995d3ed5aaba0565d71e1834604819ff9c17f5e9d5dd078f70beaf8f588b541507fed6a642c5ab42dfdf8120a7f639de5122d47a69a8e8d1"
	pubKey, err := hex.DecodeString(pubKeyHex)
	s.Require().NoError(err)

	testSetup := func() {
		sophonInfo, err := s.keeper.CreateSophonInfo(s.ctx, pubKey, types.SophonInputs{
			OwnerAddress: adminAddress,
			AdminAddress: adminAddress,
			Address:      adminAddress,
			PublicKey:    pubKey,
			Memo:         "",
		})
		s.Require().NoError(err)

		err = s.keeper.SetSophonUser(s.ctx, sophonInfo.Id, "user1", types.SophonUser{
			UserId:  "user1",
			Credits: math.NewInt(100),
		})
		s.Require().NoError(err)

		err = s.keeper.SetSophonUser(s.ctx, sophonInfo.Id, "user2", types.SophonUser{
			UserId:  "user2",
			Credits: math.NewInt(0),
		})
		s.Require().NoError(err)
	}

	s.Run("Happy path user with credits", func() {
		testSetup()

		res, err := s.queryClient.SophonEligibility(s.ctx, &types.QuerySophonEligibilityRequest{
			// 100:user1:signature_hex
			Payload: "MTAwOnVzZXIxOjZjOWIwOTNlNDM5NjdkYjZhZDUxOTM5MWNhOTFmODU1NWRkYjYyMzAzYmI4NDgzMDI4YWQ2MzgyZTViNzBiZGY1NzA1YzAwZjBhM2RlMmMyNzQ3MmE1ZWNhMzIwZDE5YWE2NzBjNTdkMzlkMWNjODE4N2RkY2ZlZjc0ZGI3ODMwMDE=",
		})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().Equal(true, res.Eligible)
		s.Require().Equal(math.NewInt(100), res.UserCredits)
		s.Require().Equal(uint64(0), res.BlockHeight)
	})

	s.Run("Happy path user without credits", func() {
		testSetup()

		res, err := s.queryClient.SophonEligibility(s.ctx, &types.QuerySophonEligibilityRequest{
			// 900:user2:signature_hex
			Payload: "OTAwOnVzZXIyOmFmNGFkY2RlYjZiMjRlMzVjZDFiYTVhZGE4YmEyOTMzMmY3Njk5NzAwZmEyZWFiMTFiYjE0M2Q2MmQzMWY5YzMxZDM5ZDIxZDlhYmQyZWQ1MTJhNThhMDkwZDU3MmU3YTFkN2RlNjQ1YmI1OTJkYjhjZWE3YjI5MWM0NjJhNWExMDA=",
		})
		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().Equal(true, res.Eligible)
		s.Require().Equal(math.NewInt(0), res.UserCredits)
		s.Require().Equal(uint64(0), res.BlockHeight)
	})

	tests := []struct {
		name        string
		wantErr     error
		errorString string
		payload     string
	}{
		{
			name: "Unknown user",
			// 987654312:user3:signature_hex
			payload: "OTg3NjU0MzEyOnVzZXIzOjU5NDkyYTdkZWUwMTdjMTMzNDI4NzQ5YjMyNWE5NzZkOTM5YzljMjJmMWU1ZGIzNjc2NDQzMGQwMzBlNGM2M2EwNTFlOTBlYTE1NWY0NjI2OTI5ZDdkYzM4NTYxMTMwMzNkNjcwYjc5YmRjYWMxOTY1MmQ4Y2Y0YmQyM2UxNWRlMDA=",
			wantErr: sdkerrors.ErrNotFound,
		},
		{
			name: "Unknown sophon",
			// 888:user1:signature_hex
			payload: "ODg4OnVzZXIxOmFhOWRkODc5NGM0NTg0ZTI1NWIwZmJmMmNmY2EyMmE5NDhhNWY0ZmY1N2IzZGQwZDhhYzdiY2Q1N2NjNjIwZmIyNWE1OGQyOTE0MDM1ZDIzNDEwOGVmNjZjY2Y4MjYzMDE2ZjZkMjhjZDZjNTRiNWU4ZjhmMDA4OWQ2ZjE1NzIxMDA=",
			wantErr: sdkerrors.ErrNotFound,
		},
		{
			name:        "Invalid payload",
			payload:     "99()()",
			wantErr:     sdkerrors.ErrInvalidRequest,
			errorString: "invalid base64 in payload",
		},
		{
			name: "Invalid number of parts",
			// 888:signature_hex
			payload:     "ODg4OmRiNGMwMWZjMzhjYTNmMzIwNTVmY2FkNGYzM2Q3YmY3ZWFhM2FmYjljMTFkNzg5NmMyNDZlN2U3NDkwYTQyYjA3OTEzMjIwZTMwOTFkMjJkZWM1N2YxNjg3ZDk0ZmM0MDhmZDE0YWIzNmI4YTE2ZTUwZTUyNGVlYTZhMWMyN2I2MDE=",
			wantErr:     sdkerrors.ErrInvalidRequest,
			errorString: "invalid number of parts",
		},
		{
			name: "Invalid block height",
			// -8:user1:signature_hex
			payload:     "LTg6dXNlcjE6MmRFVFBYMTNPUGtsNCtmRlZTS051Um9YRStlVXQ4dGVKMlg3MU5ZR1Vwa3pHRElNaGEzZEZ5aWk1eDdSQmljMkYxdkNFd2VTQ2tQSmdXeFdYbEJuUkFFPQ==",
			wantErr:     sdkerrors.ErrInvalidRequest,
			errorString: "invalid block height",
		},
		{
			name:        "Invalid signature hex",
			payload:     "ODp1c2VyMTpzUy80T0E1elZoVkRKVS9xNlJPVmhJQUQzVDlFSlNmZCsxNTV1aTlBUXJNRVFISjMwMU1jK015QVozSE1DMnZZdVZURlZMVS9LTUhRclB0VGRJQ0lFd0U9",
			wantErr:     sdkerrors.ErrInvalidRequest,
			errorString: "invalid hex in signature",
		},
		{
			name: "Invalid signature",
			// Omitted recovery param from the signature
			payload:     "ODp1c2VyMTo4MWQ5MzA1OGUwOWI3OTlkYzhhNGI3ODhhNGU1ZWE4YjQ0OGFhZWU4NDFiYmU2ZmI4MDViZjBhM2E1ZmJlNjNiNWQ0ZDg2ZDBlNTIxNDZmN2ExNjcxZGIxOGMwZTYxNmUwZDk2ZDM5YzY0MDA3NTgzNDM0MjViNmQ4YjVhNzg3Yw==",
			wantErr:     sdkerrors.ErrInvalidRequest,
			errorString: "invalid signature",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			testSetup()
			res, err := s.queryClient.SophonEligibility(s.ctx, &types.QuerySophonEligibilityRequest{
				Payload: tt.payload,
			})
			s.Require().Nil(res)
			s.Require().ErrorIs(err, tt.wantErr)

			if tt.errorString != "" {
				s.Require().Contains(err.Error(), tt.errorString)
			}
		})
	}
}
