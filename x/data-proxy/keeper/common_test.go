package keeper_test

import (
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
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	dataproxy "github.com/sedaprotocol/seda-chain/x/data-proxy"
	"github.com/sedaprotocol/seda-chain/x/data-proxy/keeper"
	"github.com/sedaprotocol/seda-chain/x/data-proxy/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	keeper      *keeper.Keeper
	bankKeeper  *testutil.MockBankKeeper
	cdc         codec.Codec
	msgSrvr     types.MsgServer
	queryClient types.QueryClient
	serverCtx   *server.Context
	authority   string
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
	encCfg := moduletestutil.MakeTestEncodingConfig(dataproxy.AppModuleBasic{})
	types.RegisterInterfaces(encCfg.InterfaceRegistry)

	ctrl := gomock.NewController(t)
	s.bankKeeper = testutil.NewMockBankKeeper(ctrl)

	s.keeper = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		s.bankKeeper,
		s.authority,
	)
	// Testvectors are generated for seda-1-testvectors
	s.ctx = testCtx.Ctx.WithChainID("seda-1-testvectors")
	s.cdc = encCfg.Codec
	s.serverCtx = server.NewDefaultContext()

	msr := keeper.NewMsgServerImpl(*s.keeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, encCfg.InterfaceRegistry)
	querier := keeper.Querier{Keeper: *s.keeper}
	types.RegisterQueryServer(queryHelper, querier)
	s.queryClient = types.NewQueryClient(queryHelper)

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *KeeperTestSuite) NewIntFromString(val string) math.Int {
	amount, success := math.NewIntFromString(val)
	s.Require().True(success)
	return amount
}

func (s *KeeperTestSuite) NewFeeFromString(val string) *sdk.Coin {
	return &sdk.Coin{
		Denom:  "aseda",
		Amount: s.NewIntFromString(val),
	}
}
