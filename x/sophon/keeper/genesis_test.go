package keeper_test

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/sophon/types"
	"github.com/stretchr/testify/require"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	gs := types.DefaultGenesisState()
	err := types.ValidateGenesis(*gs)
	require.NoError(s.T(), err)

	keeper := s.keeper
	keeper.InitGenesis(s.ctx, *gs)
}

func (s *KeeperTestSuite) TestExportImportGenesis() {
	// Set up the keeper with some state
	for sophonId := range 5 {
		sophonInfo, err := s.keeper.CreateSophonInfo(s.ctx, fmt.Appendf(nil, "sophon_info_%d", sophonId), mockSophonInfo(sophonId))
		s.Require().NoError(err)

		for user := range 10 {
			sophonUser := mockSophonUser(user)
			err := s.keeper.SetSophonUser(s.ctx, sophonInfo.Id, sophonUser.UserId, sophonUser)
			s.Require().NoError(err)
		}
	}

	for transfer := range 2 {
		err := s.keeper.SetSophonTransfer(s.ctx, uint64(transfer), mockSophonTransferAddress(transfer))
		s.Require().NoError(err)
	}

	sophonIdBeforeExport, err := s.keeper.GetCurrentSophonID(s.ctx)
	s.Require().NoError(err)

	exportedGenesis := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotNil(exportedGenesis)

	err = types.ValidateGenesis(exportedGenesis)
	s.Require().NoError(err)

	// Reset the keeper to the default state
	s.keeper.InitGenesis(s.ctx, *types.DefaultGenesisState())
	sophonIdAfterInit, err := s.keeper.GetCurrentSophonID(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(sophonIdAfterInit, collections.DefaultSequenceStart)
	hasSophonInfo, err := s.keeper.HasSophonInfo(s.ctx, fmt.Appendf(nil, "public_key_%d", 0))
	s.Require().NoError(err)
	s.Require().False(hasSophonInfo)

	s.keeper.InitGenesis(s.ctx, exportedGenesis)

	sophonIdAfterImport, err := s.keeper.GetCurrentSophonID(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(sophonIdBeforeExport, sophonIdAfterImport)

	for sophonId := range 5 {
		sophonInfo, err := s.keeper.GetSophonInfo(s.ctx, fmt.Appendf(nil, "sophon_info_%d", sophonId))
		expectedSophonInputs := mockSophonInfo(sophonId)
		s.Require().NoError(err)
		s.Require().Equal(types.SophonInfo{
			Id:           uint64(sophonId),
			OwnerAddress: expectedSophonInputs.OwnerAddress,
			AdminAddress: expectedSophonInputs.AdminAddress,
			PublicKey:    expectedSophonInputs.PublicKey,
			Address:      expectedSophonInputs.Address,
			Memo:         expectedSophonInputs.Memo,
			Balance:      expectedSophonInputs.Balance,
			UsedCredits:  expectedSophonInputs.UsedCredits,
		}, sophonInfo)

		for user := range 10 {
			sophonUser, err := s.keeper.GetSophonUser(s.ctx, uint64(sophonId), fmt.Sprintf("user_%d", user))
			s.Require().NoError(err)
			s.Require().Equal(mockSophonUser(user), sophonUser)
		}
	}

	for transfer := range 2 {
		transferAddress, err := s.keeper.GetSophonTransfer(s.ctx, uint64(transfer))
		s.Require().NoError(err)
		s.Require().Equal(mockSophonTransferAddress(transfer), transferAddress)
	}
}

func mockSophonInfo(sophonId int) types.SophonInputs {
	ownerAddr := sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "sophon_info_%d", sophonId)).PubKey().Address())
	adminAddr := sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "sophon_info_%d", sophonId)).PubKey().Address())
	address := sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "sophon_info_%d", sophonId)).PubKey().Address())

	return types.SophonInputs{
		OwnerAddress: ownerAddr.String(),
		AdminAddress: adminAddr.String(),
		PublicKey:    fmt.Appendf(nil, "public_key_%d", sophonId),
		Address:      address.String(),
		Memo:         fmt.Sprintf("memo_%d", sophonId),
		Balance:      math.NewInt(int64(sophonId+2) * 1000000000000000000),
		UsedCredits:  math.NewInt(int64(sophonId+1) * 1000000000000000000),
	}
}

func mockSophonUser(userId int) types.SophonUser {
	return types.SophonUser{
		UserId:  fmt.Sprintf("user_%d", userId),
		Credits: math.NewInt(int64(userId+1) * 100),
	}
}

func mockSophonTransferAddress(ownerNum int) []byte {
	return sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "sophon_transfer_%d", ownerNum)).PubKey().Address()).Bytes()
}
