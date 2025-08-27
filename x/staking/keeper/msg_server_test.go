package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	gomock2 "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	sdkkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	sdkstakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	sdktypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	sedatypes "github.com/sedaprotocol/seda-chain/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking"
	"github.com/sedaprotocol/seda-chain/x/staking/keeper"
	stakingtestutil "github.com/sedaprotocol/seda-chain/x/staking/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

var (
	bondedAcc     = authtypes.NewEmptyModuleAccount(sdktypes.BondedPoolName)
	notBondedAcc  = authtypes.NewEmptyModuleAccount(sdktypes.NotBondedPoolName)
	createValDesc = sdktypes.Description{
		Moniker:         "NewValidator",
		Identity:        "xyz",
		Website:         "xyz.com",
		SecurityContact: "xyz@gmail.com",
		Details:         "details",
	}
	commissionRates = sdktypes.CommissionRates{
		Rate:          math.LegacyNewDecWithPrec(5, 1),
		MaxRate:       math.LegacyNewDecWithPrec(5, 1),
		MaxChangeRate: math.LegacyNewDec(0),
	}
	minSelfDelegation = math.NewInt(1)
	value             = sdk.NewInt64Coin("aseda", 10000)
	pubKeys           = simtestutil.CreateTestPubKeys(10)
	sedaPubKeys       = generatePubKeys(10)
)

func generatePubKeys(num int) [][]byte {
	var pubKeys [][]byte
	for i := 0; i < num; i++ {
		privKey, err := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
		if err != nil {
			panic(fmt.Sprintf("failed to generate secp256k1 private key: %v", err))
		}
		pubKeys = append(pubKeys, elliptic.Marshal(privKey.PublicKey, privKey.X, privKey.Y))
	}
	return pubKeys
}

func pubKeyToAny(t *testing.T, pubKey cryptotypes.PubKey) *codectypes.Any {
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		t.Error(err)
	}
	return pkAny
}

type MsgServerTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	stakingKeeper *keeper.Keeper
	bankKeeper    *sdkstakingtestutil.MockBankKeeper
	accountKeeper *sdkstakingtestutil.MockAccountKeeper
	pubkeyKeeper  *stakingtestutil.MockPubKeyKeeper
	queryClient   sdktypes.QueryClient
	msgServer     types.MsgServer
	router        *baseapp.MsgServiceRouter
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupSuite() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	config.Seal()
}

func (s *MsgServerTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(sdktypes.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(staking.AppModuleBasic{})

	ctrl := gomock.NewController(s.T())
	pubKeyKeeper := stakingtestutil.NewMockPubKeyKeeper(ctrl)

	ctrl2 := gomock2.NewController(s.T())
	accountKeeper := sdkstakingtestutil.NewMockAccountKeeper(ctrl2)
	accountKeeper.EXPECT().GetModuleAddress(sdktypes.BondedPoolName).Return(bondedAcc.GetAddress())
	accountKeeper.EXPECT().GetModuleAddress(sdktypes.NotBondedPoolName).Return(notBondedAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("seda")).AnyTimes()

	bankKeeper := sdkstakingtestutil.NewMockBankKeeper(ctrl2)

	sdkStakingKeeper := sdkkeeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		address.NewBech32Codec("sedavaloper"),
		address.NewBech32Codec("sedavalcons"),
	)
	stakingKeeper := keeper.NewKeeper(
		sdkStakingKeeper,
		address.NewBech32Codec("sedavaloper"),
	)
	stakingKeeper.SetPubKeyKeeper(pubKeyKeeper)

	defaultParams := sdktypes.DefaultParams()
	defaultParams.BondDenom = params.DefaultBondDenom
	err := stakingKeeper.SetParams(ctx, defaultParams)
	s.Require().NoError(err)

	s.ctx = ctx
	s.stakingKeeper = stakingKeeper
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.pubkeyKeeper = pubKeyKeeper

	sdktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	sdktypes.RegisterQueryServer(queryHelper, sdkkeeper.Querier{Keeper: sdkStakingKeeper})
	s.queryClient = sdktypes.NewQueryClient(queryHelper)

	sdkMsgServer := sdkkeeper.NewMsgServerImpl(sdkStakingKeeper)
	msgServer := keeper.NewMsgServerImpl(sdkMsgServer, stakingKeeper)

	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	sdktypes.RegisterMsgServer(router, msgServer)
	types.RegisterMsgServer(router, msgServer)
	s.msgServer = msgServer
	s.router = router
}

func (s *MsgServerTestSuite) TestMsgServer() {
	testCases := []struct {
		name                   string
		input                  sdk.Msg
		address                cryptotypes.Address
		provingSchemeActivated bool
		expErr                 bool
		expErrMsg              string
	}{
		{
			name: "happy path without proving scheme",
			input: &types.MsgCreateSEDAValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[0].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[0]),
				Value:             value,
			},
			address:                pubKeys[0].Address(),
			provingSchemeActivated: false,
			expErr:                 false,
		},
		{
			name: "happy path with proving scheme",
			input: &types.MsgCreateSEDAValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[1].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[1]),
				Value:             value,
				IndexedPubKeys: []pubkeytypes.IndexedPubKey{
					{Index: 0, PubKey: sedaPubKeys[0]},
				},
			},
			address:                pubKeys[1].Address(),
			provingSchemeActivated: true,
			expErr:                 false,
		},
		{
			name: "no SEDA keys provided despite proving scheme activation",
			input: &types.MsgCreateSEDAValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[2].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[2]),
				Value:             value,
			},
			address:                pubKeys[2].Address(),
			provingSchemeActivated: true,
			expErr:                 true,
			expErrMsg:              "SEDA public keys are required",
		},
		{
			name: "SEDA keys provided despite proving scheme not activated",
			input: &types.MsgCreateSEDAValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[3].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[3]),
				Value:             value,
				IndexedPubKeys: []pubkeytypes.IndexedPubKey{
					{Index: 0, PubKey: sedaPubKeys[0]},
				},
			},
			address:                pubKeys[3].Address(),
			provingSchemeActivated: false,
			expErr:                 false,
		},
		{
			name: "empty description",
			input: &types.MsgCreateSEDAValidator{
				Description:       sdktypes.Description{},
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[4].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[4]),
				Value:             value,
			},
			address:                pubKeys[4].Address(),
			provingSchemeActivated: false,
			expErr:                 true,
			expErrMsg:              "empty description",
		},
		{
			name: "SEDA keys provided at invalid indices",
			input: &types.MsgCreateSEDAValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[5].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[5]),
				Value:             value,
				IndexedPubKeys: []pubkeytypes.IndexedPubKey{
					{Index: 1, PubKey: sedaPubKeys[0]},
					{Index: 2, PubKey: sedaPubKeys[1]},
				},
			},
			address:                pubKeys[5].Address(),
			provingSchemeActivated: true,
			expErr:                 true,
			expErrMsg:              "invalid SEDA keys",
		},
		{
			name: "too many SEDA keys provided",
			input: &types.MsgCreateSEDAValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[6].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[6]),
				Value:             value,
				IndexedPubKeys: []pubkeytypes.IndexedPubKey{
					{Index: 0, PubKey: sedaPubKeys[0]},
					{Index: 1, PubKey: sedaPubKeys[1]},
				},
			},
			address:                pubKeys[6].Address(),
			provingSchemeActivated: true,
			expErr:                 true,
			expErrMsg:              "invalid SEDA keys",
		},
		{
			name: "SDK MsgCreateValidator",
			input: &sdktypes.MsgCreateValidator{
				Description:       createValDesc,
				Commission:        commissionRates,
				MinSelfDelegation: minSelfDelegation,
				ValidatorAddress:  sdk.ValAddress(pubKeys[7].Address()).String(),
				Pubkey:            pubKeyToAny(s.T(), pubKeys[7]),
				Value:             value,
			},
			address:                pubKeys[7].Address(),
			provingSchemeActivated: false,
			expErr:                 true,
			expErrMsg:              "not implemented",
		},
		{
			name: "SDK MsgDelegate",
			input: &sdktypes.MsgDelegate{
				DelegatorAddress: sdk.AccAddress(pubKeys[0].Address()).String(),
				ValidatorAddress: sdk.ValAddress(pubKeys[0].Address()).String(),
				Amount:           value,
			},
			address:                pubKeys[0].Address(),
			provingSchemeActivated: false,
			expErr:                 false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("seda")).AnyTimes()
			s.bankKeeper.EXPECT().DelegateCoinsFromAccountToModule(gomock.Any(), sdk.AccAddress(tc.address), sdktypes.NotBondedPoolName, gomock.Any()).AnyTimes()
			s.pubkeyKeeper.EXPECT().StoreIndexedPubKeys(gomock.Any(), sdk.ValAddress(tc.address), gomock.Any()).Return(nil).MaxTimes(1)
			if tc.provingSchemeActivated {
				s.pubkeyKeeper.EXPECT().IsProvingSchemeActivated(gomock.Any(), sedatypes.SEDAKeyIndexSecp256k1).Return(true, nil).MaxTimes(1)
			} else {
				s.pubkeyKeeper.EXPECT().IsProvingSchemeActivated(gomock.Any(), sedatypes.SEDAKeyIndexSecp256k1).Return(false, nil).MaxTimes(1)
			}

			_, err := s.router.Handler(tc.input)(s.ctx, tc.input)
			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
