package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/core"
	"github.com/sedaprotocol/seda-chain/x/core/keeper"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	keeper        keeper.Keeper
	bankKeeper    *testutil.MockBankKeeper
	stakingKeeper *testutil.MockStakingKeeper
	cdc           codec.Codec
	msgSrvr       types.MsgServer
	queryClient   types.QueryClient
	serverCtx     *server.Context
	authority     string
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupSuite() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	config.Seal()
}

func (s *KeeperTestSuite) SetupTest() {
	t := s.T()
	t.Helper()

	s.authority = authtypes.NewModuleAddress("gov").String()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(core.AppModuleBasic{})
	types.RegisterInterfaces(encCfg.InterfaceRegistry)

	ctrl := gomock.NewController(t)
	s.bankKeeper = testutil.NewMockBankKeeper(ctrl)
	s.stakingKeeper = testutil.NewMockStakingKeeper(ctrl)

	s.keeper = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		nil, // wasmStorageKeeper
		nil, // batchingKeeper
		nil, // dataProxyKeeper
		s.stakingKeeper,
		s.bankKeeper,
		nil, // wasmKeeper
		nil, // wasmViewKeeper
		s.authority,
	)

	s.ctx = testCtx.Ctx.WithChainID("seda-1-testnet")
	s.cdc = encCfg.Codec
	s.serverCtx = server.NewDefaultContext()

	msr := keeper.NewMsgServerImpl(s.keeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, encCfg.InterfaceRegistry)
	querier := keeper.Querier{Keeper: s.keeper}
	types.RegisterQueryServer(queryHelper, querier)
	s.queryClient = types.NewQueryClient(queryHelper)

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) NewIntFromString(val string) math.Int {
	amount, success := math.NewIntFromString(val)
	s.Require().True(success)
	return amount
}

func (s *KeeperTestSuite) NewCoinFromString(denom, amount string) sdk.Coin {
	return sdk.Coin{
		Denom:  denom,
		Amount: s.NewIntFromString(amount),
	}
}

func (s *KeeperTestSuite) TestMsgServer_AddToAllowlist() {
	tests := []struct {
		name      string
		msg       *types.MsgAddToAllowlist
		wantErr   error
		mockSetup func()
	}{
		{
			name: "Happy path - add new public key to allowlist",
			msg: &types.MsgAddToAllowlist{
				Sender:    s.authority,
				PublicKey: "03d92f44157c939284bb101dccea8a2fc95f71ecfd35b44573a76173e3c25c67a9",
			},
			wantErr: nil,
		},
		{
			name: "Invalid authority",
			msg: &types.MsgAddToAllowlist{
				Sender:    "seda1invalid",
				PublicKey: "03d92f44157c939284bb101dccea8a2fc95f71ecfd35b44573a76173e3c25c67a9",
			},
			wantErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "Empty public key",
			msg: &types.MsgAddToAllowlist{
				Sender:    s.authority,
				PublicKey: "",
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "Already allowlisted from first test case",
			msg: &types.MsgAddToAllowlist{
				Sender:    s.authority,
				PublicKey: "03d92f44157c939284bb101dccea8a2fc95f71ecfd35b44573a76173e3c25c67a9",
			},
			wantErr: types.ErrAlreadyAllowlisted,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			res, err := s.msgSrvr.AddToAllowlist(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(res)

			// Verify the public key was added to allowlist
			exists, err := s.keeper.IsAllowlisted(s.ctx, tt.msg.PublicKey)
			s.Require().NoError(err)
			s.Require().True(exists)
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_Stake() {
	// Valid test data
	validPublicKey := "0364c055bb57396faf8e86cb39a473177875259b83c5828d71f04d2dd101b2e935"
	validProof := "032502c5864b629b8201bfa5176026db6ba821cdc68a16e151b89405f220907b9f2a61fa03ff198b4bb59f52a6094a668b5e66ac6938c4e569f89fb74c9131ed4b4284083cd9e5791455371fefa2141485"
	validMemo := ""
	validStake := s.NewCoinFromString("aseda", "1000000000000000000") // 1 aseda
	validSender := "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5"
	validSenderAddr := sdk.MustAccAddressFromBech32(validSender)

	tests := []struct {
		name    string
		msg     *types.MsgStake
		wantErr error
		setup   func()
		postRun func()
	}{
		{
			name: "Invalid public key hex",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: "invalid_hex",
				Memo:      validMemo,
				Proof:     validProof,
				Stake:     validStake,
			},
			wantErr: hex.InvalidByteError('i'),
		},
		{
			name: "Invalid proof hex",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     "invalid_hex",
				Stake:     validStake,
			},
			wantErr: hex.InvalidByteError('i'),
		},
		{
			name: "Invalid stake proof",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     "032c74385c590d76e1a6e15364f515f0ae38ba61077c276dcf6aea4a810a36e4988a32cccfd9b08c8ab74f3e4e6dbb6f8e600364432bb166361018f45b817b350b30ae352b7131ab267dffcd643057c484", // Invalid proof
				Stake:     validStake,
			},
			wantErr: types.ErrInvalidStakeProof,
		},
		{
			name: "Invalid stake denom",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     validProof,
				Stake:     s.NewCoinFromString("uatom", "1000000"),
			},
			wantErr: sdkerrors.ErrInvalidCoins,
			setup: func() {
				s.keeper.AddToAllowlist(s.ctx, validPublicKey)
				s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil)
			},
		},
		{
			name: "Insufficient stake amount for new staker",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     validProof,
				Stake:     s.NewCoinFromString("aseda", "100"), // Below minimum
			},
			wantErr: types.ErrInsufficientStake,
			setup: func() {
				s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil)
			},
		},
		{
			name: "Invalid sender address",
			msg: &types.MsgStake{
				Sender:    "invalid_address",
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     validProof,
				Stake:     validStake,
			},
			wantErr: sdkerrors.ErrInvalidAddress,
			setup: func() {
				s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil)
			},
		},
		{
			name: "Bank keeper error - insufficient funds",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     validProof,
				Stake:     validStake,
			},
			wantErr: sdkerrors.ErrInsufficientFunds,
			setup: func() {
				s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil)
				senderAddr, _ := sdk.AccAddressFromBech32(validSender)
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(), senderAddr, types.ModuleName, sdk.NewCoins(validStake),
				).Return(sdkerrors.ErrInsufficientFunds)
			},
		},
		{
			name: "Happy path - new staker with valid proof",
			msg: &types.MsgStake{
				Sender:    validSender,
				PublicKey: validPublicKey,
				Memo:      validMemo,
				Proof:     validProof,
				Stake:     validStake,
			},
			wantErr: nil,
			setup: func() {
				// Must be allowlisted
				s.keeper.AddToAllowlist(s.ctx, validPublicKey)

				// Mock staking keeper to return bond denom
				s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil)
				// Mock bank keeper to send coins
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(), validSenderAddr, types.ModuleName, sdk.NewCoins(validStake),
				).Return(nil)
			},
			postRun: func() {
				// Second stake with different proof (sequence number 1)
				secondStake := s.NewCoinFromString("aseda", "500000000000000000")
				secondMsg := &types.MsgStake{
					Sender:    validSender,
					PublicKey: validPublicKey,
					Memo:      "VGhlIFNpbmdsZSBVTklYIFNwZWNpZmljYXRpb24gc3VwcG9ydHMgZm9ybWFsIHN0YW5kYXJkcyBkZXZlbG9wZWQgZm9yIGFwcGxpY2F0aW9ucyBwb3J0YWJpbGl0eS4g",
					Proof:     "02ed19ea7f12d53d18c93297c358a46b8d974ebeca589afb9f5063963c3d2d61835d12dae08768e006821da382bfd12b0fac66e5a4f800b73ecc128c4a6c3687507c069c44b831bcbab19f68e9ca1303ed",
					Stake:     secondStake,
				}

				s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil)
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(), validSenderAddr, types.ModuleName, sdk.NewCoins(secondStake),
				).Return(nil)

				res, err := s.msgSrvr.Stake(s.ctx, secondMsg)
				s.Require().NoError(err)
				s.Require().NotNil(res)

				// Verify total stake
				staker, err := s.keeper.GetStaker(s.ctx, validPublicKey)
				s.Require().NoError(err)
				expectedTotal := validStake.Amount.Add(secondStake.Amount)
				s.Require().Equal(expectedTotal, staker.Staked)
				s.Require().Equal(secondMsg.Memo, staker.Memo) // Memo should be updated
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.setup != nil {
				tt.setup()
			}

			res, err := s.msgSrvr.Stake(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(res)

			// Verify staker was created/updated
			staker, err := s.keeper.GetStaker(s.ctx, tt.msg.PublicKey)
			s.Require().NoError(err)
			s.Require().Equal(tt.msg.PublicKey, staker.PublicKey)
			s.Require().Equal(tt.msg.Memo, staker.Memo)
			s.Require().Equal(tt.msg.Stake.Amount, staker.Staked)

			if tt.postRun != nil {
				tt.postRun()
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgServer_UpdateParams() {
	authority := s.keeper.GetAuthority()

	tests := []struct {
		name    string
		msg     *types.MsgUpdateParams
		wantErr error
	}{
		{
			name: "Happy path - valid params update",
			msg: &types.MsgUpdateParams{
				Authority: authority,
				Params:    types.DefaultParams(),
			},
			wantErr: nil,
		},
		{
			name: "Invalid authority",
			msg: &types.MsgUpdateParams{
				Authority: "seda1invalid",
				Params: types.Params{
					StakingConfig: types.StakingConfig{
						MinimumStake:     s.NewIntFromString("1000000000000000000"),
						AllowlistEnabled: false,
					},
				},
			},
			wantErr: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "Unauthrized authority",
			msg: &types.MsgUpdateParams{
				Authority: "seda1uea9km4nup9q7qu96ak683kc67x9jf7ste45z5",
				Params: types.Params{
					StakingConfig: types.StakingConfig{
						MinimumStake:     s.NewIntFromString("1000000000000000000"),
						AllowlistEnabled: false,
					},
				},
			},
			wantErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "Invalid params",
			msg: &types.MsgUpdateParams{
				Authority: authority,
				Params: types.Params{
					StakingConfig: types.StakingConfig{
						MinimumStake:     s.NewIntFromString("-1"), // Invalid negative amount
						AllowlistEnabled: false,
					},
				},
			},
			wantErr: sdkerrors.ErrInvalidRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := s.msgSrvr.UpdateParams(s.ctx, tt.msg)
			if tt.wantErr != nil {
				s.Require().ErrorIs(err, tt.wantErr)
				s.Require().Nil(res)
				return
			}
			s.Require().NoError(err)
			s.Require().NotNil(res)

			// Verify params were updated
			params, err := s.keeper.GetParams(s.ctx)
			s.Require().NoError(err)
			s.Require().Equal(tt.msg.Params, params)
		})
	}
}
