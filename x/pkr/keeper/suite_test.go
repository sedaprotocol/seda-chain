package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sedaprotocol/seda-chain/x/pkr"

	pkrtypes "github.com/sedaprotocol/seda-chain/x/pkr/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/sedaprotocol/seda-chain/x/pkr/keeper"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	keeper      *keeper.Keeper
	cdc         codec.Codec
	msgSrvr     pkrtypes.MsgServer
	queryClient pkrtypes.QueryClient
	serverCtx   *server.Context
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	pkrKeeper, enCfg, ctx := setupKeeper(s.T())
	s.keeper = pkrKeeper
	s.ctx = ctx
	s.cdc = enCfg.Codec
	s.serverCtx = server.NewDefaultContext()

	msr := keeper.NewMsgServerImpl(*pkrKeeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enCfg.InterfaceRegistry)
	querier := keeper.Querier{Keeper: *s.keeper}
	pkrtypes.RegisterQueryServer(queryHelper, querier)
	s.queryClient = pkrtypes.NewQueryClient(queryHelper)
}

func setupKeeper(t *testing.T) (*keeper.Keeper, moduletestutil.TestEncodingConfig, sdk.Context) {
	t.Helper()
	key := storetypes.NewKVStoreKey(pkrtypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(pkr.AppModuleBasic{})
	pkrtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	pkrKeeper := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key))

	return pkrKeeper, encCfg, testCtx.Ctx
}
