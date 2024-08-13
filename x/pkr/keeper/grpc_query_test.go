package keeper_test

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gomock "go.uber.org/mock/gomock"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

func (s *KeeperTestSuite) TestQuerier_ValidatorKeys() {
	pubKeys, valAddrs := s.generatePubKeysAndValAddrs(10)

	// Store the public keys - one for each validator.
	for i := range pubKeys {
		addMsg := types.MsgAddKey{
			ValidatorAddr: valAddrs[i].String(),
			IndexedPubKeys: []types.IndexedPubKey{
				{
					Index:  uint32(i),
					PubKey: pubKeys[i],
				},
			},
		}

		// Mock GetValidator()
		s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddrs[i].Bytes()).Return(stakingtypes.Validator{}, nil)

		resp, err := s.msgSrvr.AddKey(s.ctx, &addMsg)
		s.Require().NoError(err)
		s.Require().NotNil(resp)
	}

	for j := range valAddrs {
		resp, err := s.queryClient.ValidatorKeys(s.ctx, &types.QueryValidatorKeysRequest{ValidatorAddr: valAddrs[j].String()})
		s.Require().NoError(err)

		s.Require().Equal(1, len(resp.ValidatorPubKeys.IndexedPubKeys))
		pk := resp.ValidatorPubKeys.IndexedPubKeys[0]
		s.Require().Equal(uint32(j), pk.Index)
		s.Require().Equal(pubKeys[j].Value, pk.PubKey.Value)
	}
}
