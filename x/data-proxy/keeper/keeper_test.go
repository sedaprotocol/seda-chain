package keeper_test

import (
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	"cosmossdk.io/collections"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

func (s *KeeperTestSuite) TestKeeper_EndBlockFeeUpdate() {
	s.Run("Apply single pending fee update", func() {
		s.SetupTest()

		pubKeyBytes := secp256k1.GenPrivKey().PubKey().Bytes()
		pubKeyHex := hex.EncodeToString(pubKeyBytes)

		err := s.keeper.DataProxyConfigs.Set(s.ctx, pubKeyBytes, types.ProxyConfig{
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee:           s.NewFeeFromString("10000"),
			Memo:          "test",
			FeeUpdate: &types.FeeUpdate{
				UpdateHeight: 0,
				NewFee:       *s.NewFeeFromString("987654321"),
			},
			AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		})
		s.Require().NoError(err)
		err = s.keeper.FeeUpdateQueue.Set(s.ctx, collections.Join(int64(0), pubKeyBytes))
		s.Require().NoError(err)

		err = s.keeper.EndBlock(s.ctx)
		s.Require().NoError(err)

		proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, pubKeyHex)
		s.Require().NoError(err)

		s.Require().Equal(types.ProxyConfig{
			PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			Fee:           s.NewFeeFromString("987654321"),
			Memo:          "test",
			FeeUpdate:     nil,
			AdminAddress:  "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
		}, proxyConfig)

		found, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(int64(0), pubKeyBytes))
		s.Require().NoError(err)
		s.Require().False(found)
	})

	s.Run("Apply multiple pending fee updates", func() {
		s.SetupTest()

		pubKeys := []struct {
			pubKeyBytes []byte
			pubKeyHex   string
		}{}
		for i := 0; i < 10; i++ {
			pubKeyBytes := secp256k1.GenPrivKey().PubKey().Bytes()
			pubKeyHex := hex.EncodeToString(pubKeyBytes)

			pubKeys = append(pubKeys, struct {
				pubKeyBytes []byte
				pubKeyHex   string
			}{pubKeyBytes: pubKeyBytes, pubKeyHex: pubKeyHex})

			updateHeight := int64(0)
			if i >= 5 {
				updateHeight = int64(1)
			}

			err := s.keeper.DataProxyConfigs.Set(s.ctx, pubKeyBytes, types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("10"),
				Memo:          "test",
				FeeUpdate: &types.FeeUpdate{
					UpdateHeight: updateHeight,
					NewFee:       *s.NewFeeFromString("30"),
				},
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			})
			s.Require().NoError(err)
			err = s.keeper.FeeUpdateQueue.Set(s.ctx, collections.Join(updateHeight, pubKeyBytes))
			s.Require().NoError(err)
		}

		err := s.keeper.EndBlock(s.ctx)
		s.Require().NoError(err)

		for i, testInput := range pubKeys {
			proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, testInput.pubKeyHex)
			s.Require().NoError(err)

			expected := types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("10"),
				Memo:          "test",
				FeeUpdate: &types.FeeUpdate{
					UpdateHeight: 1,
					NewFee:       *s.NewFeeFromString("30"),
				},
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			}
			if i < 5 {
				expected.Fee = &expected.FeeUpdate.NewFee
				expected.FeeUpdate = nil
			}

			s.Require().Equal(expected, proxyConfig)

			updateHeight := int64(0)
			if i >= 5 {
				updateHeight = int64(1)
			}
			found, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(updateHeight, testInput.pubKeyBytes))
			s.Require().NoError(err)
			if i < 5 {
				s.Require().False(found)
			} else {
				s.Require().True(found)
			}
		}

		s.ctx = s.ctx.WithBlockHeight(1)

		err = s.keeper.EndBlock(s.ctx)
		s.Require().NoError(err)

		for i, testInput := range pubKeys {
			proxyConfig, err := s.keeper.GetDataProxyConfig(s.ctx, testInput.pubKeyHex)
			s.Require().NoError(err)

			expected := types.ProxyConfig{
				PayoutAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Fee:           s.NewFeeFromString("10"),
				Memo:          "test",
				FeeUpdate: &types.FeeUpdate{
					UpdateHeight: 1,
					NewFee:       *s.NewFeeFromString("30"),
				},
				AdminAddress: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
			}

			expected.Fee = &expected.FeeUpdate.NewFee
			expected.FeeUpdate = nil

			s.Require().Equal(expected, proxyConfig)

			updateHeight := int64(0)
			if i >= 5 {
				updateHeight = int64(1)
			}
			found, err := s.keeper.FeeUpdateQueue.Has(s.ctx, collections.Join(updateHeight, testInput.pubKeyBytes))
			s.Require().NoError(err)
			s.Require().False(found)
		}
	})
}
