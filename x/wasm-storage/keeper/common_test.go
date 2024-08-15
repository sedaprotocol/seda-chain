package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var mockedByteArray = []byte("82a9dda829eb7f8ffe9fbe49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e5082")

type KeeperTestSuite struct {
	suite.Suite
	ctx         sdk.Context
	keeper      *keeper.Keeper
	cdc         codec.Codec
	msgSrvr     wasmstoragetypes.MsgServer
	queryClient wasmstoragetypes.QueryClient
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
	s.authority = authtypes.NewModuleAddress("gov").String()
	wasmStorageKeeper, enCfg, ctx := setupKeeper(s.T(), s.authority)
	s.keeper = wasmStorageKeeper
	s.ctx = ctx
	s.cdc = enCfg.Codec

	msr := keeper.NewMsgServerImpl(*wasmStorageKeeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, enCfg.InterfaceRegistry)
	querier := keeper.NewQuerierImpl(*s.keeper)
	wasmstoragetypes.RegisterQueryServer(queryHelper, querier)
	s.queryClient = wasmstoragetypes.NewQueryClient(queryHelper)

	err := s.keeper.Params.Set(ctx, wasmstoragetypes.DefaultParams())
	s.Require().NoError(err)
}

func setupKeeper(t *testing.T, authority string) (*keeper.Keeper, moduletestutil.TestEncodingConfig, sdk.Context) {
	t.Helper()
	key := storetypes.NewKVStoreKey(wasmstoragetypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(wasmstorage.AppModuleBasic{})
	wasmstoragetypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	wasmStorageKeeper := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), authority, nil, nil, nil, nil)

	return wasmStorageKeeper, encCfg, ctx
}
