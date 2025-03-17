package keeper_test

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkcdctestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sedaprotocol/seda-chain/app/params"
	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testutil"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var mockedByteArray = []byte("82a9dda829eb7f8ffe9fbe49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e5082")

type KeeperTestSuite struct {
	suite.Suite
	ctx               sdk.Context
	keeper            *keeper.Keeper
	cdc               codec.Codec
	msgSrvr           wasmstoragetypes.MsgServer
	queryClient       wasmstoragetypes.QueryClient
	authority         string
	txConfig          client.TxConfig
	mockBankKeeper    *testutil.MockBankKeeper
	mockStakingKeeper *testutil.MockStakingKeeper
	mockWasmKeeper    *testutil.MockContractOpsKeeper
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
	aminoCodec := codec.NewLegacyAmino()
	interfaceRegistry := sdkcdctestutil.CodecOptions{
		AccAddressPrefix: params.Bech32PrefixAccAddr,
		ValAddressPrefix: params.Bech32PrefixValAddr,
	}.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)

	encCfg := moduletestutil.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             aminoCodec,
	}

	mb := module.NewBasicManager(wasmstorage.AppModuleBasic{}, wasm.AppModuleBasic{})

	std.RegisterLegacyAminoCodec(encCfg.Amino)
	std.RegisterInterfaces(encCfg.InterfaceRegistry)
	mb.RegisterLegacyAminoCodec(encCfg.Amino)
	mb.RegisterInterfaces(encCfg.InterfaceRegistry)

	s.txConfig = encCfg.TxConfig
	s.authority = authtypes.NewModuleAddress("gov").String()
	s.ctx = s.SetupKeepers(s.T(), s.authority, encCfg)

	msr := keeper.NewMsgServerImpl(*s.keeper)
	s.msgSrvr = msr

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, encCfg.InterfaceRegistry)
	querier := keeper.NewQuerierImpl(*s.keeper)
	wasmstoragetypes.RegisterQueryServer(queryHelper, querier)
	s.queryClient = wasmstoragetypes.NewQueryClient(queryHelper)

	err := s.keeper.Params.Set(s.ctx, wasmstoragetypes.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) SetupKeepers(t *testing.T, authority string, encCfg moduletestutil.TestEncodingConfig) sdk.Context {
	t.Helper()
	key := storetypes.NewKVStoreKey(wasmstoragetypes.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	ctrl := gomock.NewController(t)
	mockBankKeeper := testutil.NewMockBankKeeper(ctrl)
	mockStakingKeeper := testutil.NewMockStakingKeeper(ctrl)
	mockWasmKeeper := testutil.NewMockContractOpsKeeper(ctrl)

	wasmStorageKeeper := keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(key),
		authority,
		authtypes.FeeCollectorName,
		encCfg.TxConfig.TxDecoder(),
		mockBankKeeper,
		mockStakingKeeper,
		mockWasmKeeper,
	)

	s.mockBankKeeper = mockBankKeeper
	s.mockStakingKeeper = mockStakingKeeper
	s.mockWasmKeeper = mockWasmKeeper
	s.keeper = wasmStorageKeeper

	return ctx
}

func (s *KeeperTestSuite) ApplyDefaultMockExpectations() {
	s.mockBankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), authtypes.FeeCollectorName, gomock.Any()).Return(nil).AnyTimes()
	s.mockStakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("aseda", nil).AnyTimes()
}
