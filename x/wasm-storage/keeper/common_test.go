package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var mockedByteArray = []byte("82a9dda829eb7f8ffe9fbe49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e5082")
var mockedByteArray2 = []byte("a9dda829eb7f8ffe9fbesfa49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e50829a")

type KeeperTestSuite struct {
	suite.Suite
	ctx               sdk.Context
	wasmStorageKeeper *keeper.Keeper
	blockTime         time.Time
	cdc               codec.Codec
	msgSrvr           wasmstoragetypes.MsgServer
	queryClient       wasmstoragetypes.QueryClient
	authority         string
}

func (s *KeeperTestSuite) SetupTest() {
	s.authority = authtypes.NewModuleAddress("gov").String()
	wasmStorageKeeper, enCfg, ctx := setupKeeper(s.T(), s.authority)
	s.wasmStorageKeeper = wasmStorageKeeper
	s.ctx = ctx
	s.cdc = enCfg.Codec

	msr := keeper.NewMsgServerImpl(*wasmStorageKeeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enCfg.InterfaceRegistry)
	querier := keeper.NewQuerierImpl(*s.wasmStorageKeeper)
	wasmstoragetypes.RegisterQueryServer(queryHelper, querier)
	s.queryClient = wasmstoragetypes.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func setupKeeper(t *testing.T, authority string) (*keeper.Keeper, moduletestutil.TestEncodingConfig, sdk.Context) {
	t.Helper()
	key := sdk.NewKVStoreKey(wasmstoragetypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(wasmstorage.AppModuleBasic{})
	wasmstoragetypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	wasmStorageKeeper := keeper.NewKeeper(encCfg.Codec, key, authority, nil)

	return wasmStorageKeeper, encCfg, ctx
}
