package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/pkr"
	"github.com/sedaprotocol/seda-chain/x/pkr/keeper"
	"github.com/sedaprotocol/seda-chain/x/pkr/keeper/testutil"
	pkrtypes "github.com/sedaprotocol/seda-chain/x/pkr/types"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx               sdk.Context
	keeper            *keeper.Keeper
	mockStakingKeeper *testutil.MockStakingKeeper
	cdc               codec.Codec
	valCdc            address.Codec
	msgSrvr           pkrtypes.MsgServer
	queryClient       pkrtypes.QueryClient
	serverCtx         *server.Context
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	t := s.T()
	t.Helper()
	key := storetypes.NewKVStoreKey(pkrtypes.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(pkr.AppModuleBasic{})
	pkrtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	ctrl := gomock.NewController(t)
	s.mockStakingKeeper = testutil.NewMockStakingKeeper(ctrl)
	s.valCdc = addresscodec.NewBech32Codec(params.Bech32PrefixValAddr)
	s.keeper = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		s.mockStakingKeeper,
		s.valCdc,
	)
	s.ctx = testCtx.Ctx
	s.cdc = encCfg.Codec
	s.serverCtx = server.NewDefaultContext()

	msr := keeper.NewMsgServerImpl(*s.keeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, encCfg.InterfaceRegistry)
	querier := keeper.Querier{Keeper: *s.keeper}
	pkrtypes.RegisterQueryServer(queryHelper, querier)
	s.queryClient = pkrtypes.NewQueryClient(queryHelper)
}
