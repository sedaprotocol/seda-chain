package keeper_test

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/fast/types"
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
	for fastClientID := range 5 {
		fastClient, err := s.keeper.CreateFastClient(s.ctx, fmt.Appendf(nil, "fast_client_%d", fastClientID), mockFastClient(fastClientID))
		s.Require().NoError(err)

		for user := range 10 {
			fastUser := mockFastUser(user)
			err := s.keeper.SetFastUser(s.ctx, fastClient.Id, fastUser.UserId, fastUser)
			s.Require().NoError(err)
		}
	}

	for transfer := range 2 {
		err := s.keeper.SetFastTransfer(s.ctx, uint64(transfer), mockFastClientTransferAddress(transfer))
		s.Require().NoError(err)
	}

	fastClientIDBeforeExport, err := s.keeper.GetCurrentFastClientID(s.ctx)
	s.Require().NoError(err)

	exportedGenesis := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotNil(exportedGenesis)

	err = types.ValidateGenesis(exportedGenesis)
	s.Require().NoError(err)

	// Reset the keeper to the default state
	s.keeper.InitGenesis(s.ctx, *types.DefaultGenesisState())
	fastClientIDAfterInit, err := s.keeper.GetCurrentFastClientID(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(fastClientIDAfterInit, collections.DefaultSequenceStart)
	hasFastClient, err := s.keeper.HasFastClient(s.ctx, fmt.Appendf(nil, "public_key_%d", 0))
	s.Require().NoError(err)
	s.Require().False(hasFastClient)

	s.keeper.InitGenesis(s.ctx, exportedGenesis)

	fastClientIDAfterImport, err := s.keeper.GetCurrentFastClientID(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(fastClientIDBeforeExport, fastClientIDAfterImport)

	for fastClientID := range 5 {
		fastClient, err := s.keeper.GetFastClient(s.ctx, fmt.Appendf(nil, "fast_client_%d", fastClientID))
		expectedFastClientInputs := mockFastClient(fastClientID)
		s.Require().NoError(err)
		s.Require().Equal(types.FastClient{
			Id:           uint64(fastClientID),
			OwnerAddress: expectedFastClientInputs.OwnerAddress,
			AdminAddress: expectedFastClientInputs.AdminAddress,
			PublicKey:    expectedFastClientInputs.PublicKey,
			Address:      expectedFastClientInputs.Address,
			Memo:         expectedFastClientInputs.Memo,
			Balance:      expectedFastClientInputs.Balance,
			UsedCredits:  expectedFastClientInputs.UsedCredits,
		}, fastClient)

		for user := range 10 {
			fastUser, err := s.keeper.GetFastUser(s.ctx, uint64(fastClientID), fmt.Sprintf("user_%d", user))
			s.Require().NoError(err)
			s.Require().Equal(mockFastUser(user), fastUser)
		}
	}

	for transfer := range 2 {
		transferAddress, err := s.keeper.GetFastTransfer(s.ctx, uint64(transfer))
		s.Require().NoError(err)
		s.Require().Equal(mockFastClientTransferAddress(transfer), transferAddress)
	}
}

func mockFastClient(fastClientID int) types.FastClientInput {
	ownerAddr := sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "fast_client_%d", fastClientID)).PubKey().Address())
	adminAddr := sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "fast_client_%d", fastClientID)).PubKey().Address())
	address := sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "fast_client_%d", fastClientID)).PubKey().Address())

	return types.FastClientInput{
		OwnerAddress: ownerAddr.String(),
		AdminAddress: adminAddr.String(),
		PublicKey:    fmt.Appendf(nil, "public_key_%d", fastClientID),
		Address:      address.String(),
		Memo:         fmt.Sprintf("memo_%d", fastClientID),
		Balance:      math.NewInt(int64(fastClientID+2) * 1000000000000000000),
		UsedCredits:  math.NewInt(int64(fastClientID+1) * 1000000000000000000),
	}
}

func mockFastUser(userId int) types.FastUser {
	return types.FastUser{
		UserId:  fmt.Sprintf("user_%d", userId),
		Credits: math.NewInt(int64(userId+1) * 100),
	}
}

func mockFastClientTransferAddress(ownerNum int) []byte {
	return sdk.AccAddress(ed25519.GenPrivKeyFromSecret(fmt.Appendf(nil, "fast_client_transfer_%d", ownerNum)).PubKey().Address()).Bytes()
}
