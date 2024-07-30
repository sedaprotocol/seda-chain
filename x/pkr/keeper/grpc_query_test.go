package keeper_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gomock "go.uber.org/mock/gomock"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func (s *KeeperTestSuite) TestQuerier_ValidatorKeys() {
	keyFileDir := filepath.Dir(s.serverCtx.Config.PrivValidatorKeyFile())
	pubKeys := make([]cryptotypes.PubKey, 0, 3)
	pubKeysAny := make([]*codectypes.Any, 0, 3)
	for i := 0; i < 3; i++ {
		pk, err := utils.LoadOrGenVRFKey(s.serverCtx.Config, "", "")
		s.Require().NoError(err)
		pubKeys = append(pubKeys, pk)

		pkAny, err := codectypes.NewAnyWithValue(pk)
		s.Require().NoError(err)
		pubKeysAny = append(pubKeysAny, pkAny)

		// Remove key file so we can create another one.
		s.Require().NoError(os.Remove(filepath.Join(keyFileDir, utils.VRFKeyFileName)))
	}

	s.T().Cleanup(func() {
		s.Require().NoError(os.RemoveAll(keyFileDir))
	})

	// Store the pubKeys
	valAddr := "sedavaloper10hpwdkc76wgqm5lg4my6vz33kps0jr05u9uxga"
	for i, pk := range pubKeysAny {
		addMsg := types.MsgAddKey{
			ValidatorAddress: valAddr,
			Index:            uint32(i),
			Pubkey:           pk,
		}

		// Validator must exist.
		s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{}, nil)

		resp, err := s.msgSrvr.AddKey(s.ctx, &addMsg)
		s.Require().NoError(err)
		s.Require().NotNil(resp)
	}

	resp, err := s.queryClient.ValidatorKeys(s.ctx, &types.QueryValidatorKeysRequest{ValidatorAddr: valAddr})
	s.Require().NoError(err)
	s.Require().Equal(len(pubKeys), len(resp.IndexPubkeyPairs))
	for i := range pubKeys {
		expected := strings.ToUpper(hex.EncodeToString(pubKeys[i].Bytes()))
		s.Require().True(resp.IndexPubkeyPairs[i] == fmt.Sprintf("%d,PubKeySecp256k1{%s}", i, expected))
	}
}
