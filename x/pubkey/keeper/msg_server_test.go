package keeper_test

import (
	"encoding/hex"
	"os"
	"path/filepath"

	gomock "go.uber.org/mock/gomock"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

func (s *KeeperTestSuite) TestMsgServer_InitialAddKey() {
	pubKeys, valAddrs := s.generatePubKeysAndValAddrs(10)

	tests := []struct {
		name       string
		msg        *types.MsgAddKey
		valAddr    sdk.ValAddress
		wantErr    error
		wantErrMsg string
	}{
		{
			name: "Happy path",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[0].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index:  0,
						PubKey: pubKeys[0],
					},
				},
			},
			valAddr: valAddrs[0],
			wantErr: nil,
		},
		{
			name: "Empty validator address",
			msg: &types.MsgAddKey{
				ValidatorAddr: "",
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index:  0,
						PubKey: pubKeys[0],
					},
				},
			},
			valAddr:    valAddrs[0],
			wantErr:    sdkerrors.ErrInvalidRequest,
			wantErrMsg: "empty validator address",
		},
		{
			name: "Too many",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[1].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index:  12,
						PubKey: pubKeys[0],
					},
					{
						Index:  48,
						PubKey: pubKeys[1],
					},
				},
			},
			valAddr:    valAddrs[1],
			wantErr:    sdkerrors.ErrInvalidRequest,
			wantErrMsg: "invalid number of SEDA keys",
		},
		{
			name: "Incorrect index",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[1].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index:  3,
						PubKey: pubKeys[0],
					},
				},
			},
			valAddr:    valAddrs[1],
			wantErr:    sdkerrors.ErrInvalidRequest,
			wantErrMsg: "invalid SEDA key index",
		},
		{
			name: "Incorrect pubkey format",
			msg: &types.MsgAddKey{
				ValidatorAddr: valAddrs[2].String(),
				IndexedPubKeys: []types.IndexedPubKey{
					{
						Index: 0,
						PubKey: func() []byte {
							return secp256k1.GenPrivKey().PubKey().Bytes()
						}(),
					},
				},
			},
			valAddr:    valAddrs[2],
			wantErr:    sdkerrors.ErrInvalidRequest,
			wantErrMsg: "invalid public key at SEDA key index",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Cleanup(func() {
				path := s.serverCtx.Config.PrivValidatorKeyFile()
				path = filepath.Dir(path)
				s.Require().NoError(os.RemoveAll(path))
			})

			// Mock validator store.
			s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), tt.valAddr.Bytes()).Return(stakingtypes.Validator{}, nil).AnyTimes()

			got, err := s.msgSrvr.AddKey(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Contains(err.Error(), tt.wantErrMsg)
				s.Require().Nil(got)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(got)
			// Check the validator's keys at once.
			pksActual, err := s.keeper.GetValidatorKeys(s.ctx, tt.valAddr.String())
			s.Require().NoError(err)
			s.Require().Equal(tt.valAddr.String(), pksActual.ValidatorAddr)
			s.Require().Equal(len(tt.msg.IndexedPubKeys), len(pksActual.IndexedPubKeys))

			// Check each index.
			for _, indPubKey := range tt.msg.IndexedPubKeys {
				pkActual, err := s.keeper.GetValidatorKeyAtIndex(s.ctx, tt.valAddr, sedatypes.SEDAKeyIndex(indPubKey.Index))
				s.Require().NoError(err)
				s.Require().Equal(indPubKey.PubKey, pkActual)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_SubsequentAddKey() {
	s.Run("AddKey should return previous keys registered for the validator if present", func() {
		pubKeys, valAddrs := s.generatePubKeysAndValAddrs(3)

		// Mock validator store.
		valAddr := valAddrs[0]
		s.mockStakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr.Bytes()).Return(stakingtypes.Validator{}, nil).AnyTimes()

		// First add key message
		firstKey := pubKeys[0]
		msg1 := &types.MsgAddKey{
			ValidatorAddr: valAddr.String(),
			IndexedPubKeys: []types.IndexedPubKey{
				{
					Index:  0,
					PubKey: firstKey,
				},
			},
		}

		got1, err := s.msgSrvr.AddKey(s.ctx, msg1)
		s.Require().NoError(err)
		s.Require().NotNil(got1)

		// Second add key message with different key
		secondKey := pubKeys[1]
		msg2 := &types.MsgAddKey{
			ValidatorAddr: valAddr.String(),
			IndexedPubKeys: []types.IndexedPubKey{
				{
					Index:  0,
					PubKey: secondKey,
				},
			},
		}

		got2, err := s.msgSrvr.AddKey(s.ctx, msg2)
		s.Require().NoError(err)
		s.Require().NotNil(got2)
		// Check that we emitted the previously registered keys.
		events := s.ctx.EventManager().Events()
		// 3 events: 1 for the first key, 1 for the new key, 1 for the removal of the first key
		s.Require().Len(events, 3)
		s.Require().Equal(events[2].Type, types.EventTypeRemoveKey)
		s.Require().Equal(events[2].Attributes[0].Key, types.AttributeValidatorAddr)
		s.Require().Equal(events[2].Attributes[0].Value, valAddr.String())
		s.Require().Equal(events[2].Attributes[1].Key, types.AttributePubKeyIndex)
		s.Require().Equal(events[2].Attributes[1].Value, "0")
		s.Require().Equal(events[2].Attributes[2].Key, types.AttributePublicKey)
		s.Require().Equal(events[2].Attributes[2].Value, hex.EncodeToString(firstKey))

		thirdKey := pubKeys[2]
		msg3 := &types.MsgAddKey{
			ValidatorAddr: valAddr.String(),
			IndexedPubKeys: []types.IndexedPubKey{
				{
					Index:  0,
					PubKey: thirdKey,
				},
			},
		}

		got3, err := s.msgSrvr.AddKey(s.ctx, msg3)
		s.Require().NoError(err)
		s.Require().NotNil(got3)
		// 5 events: the 3 previous events + 1 for the third key and 1 for the removal of the second key
		s.Require().Len(s.ctx.EventManager().Events(), 5)
		s.Require().Equal(s.ctx.EventManager().Events()[4].Attributes[2].Value, hex.EncodeToString(secondKey))
	})
}
