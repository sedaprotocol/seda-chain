package integration

import (
	"fmt"
	"time"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

const appName = "integration-app"

// IntegationApp is a test application that can be used to test the integration of modules.
type IntegationApp struct {
	*baseapp.BaseApp

	ctx           sdk.Context
	logger        log.Logger
	moduleManager module.Manager
	queryHelper   *baseapp.QueryServiceTestHelper
}

// NewIntegrationApp creates an application for testing purposes. This application
// is able to route messages to their respective handlers.
func NewIntegrationApp(
	sdkCtx sdk.Context,
	logger log.Logger,
	keys map[string]*storetypes.KVStoreKey,
	appCodec codec.Codec,
	router *baseapp.MsgServiceRouter,
	modules map[string]appmodule.AppModule,
) *IntegationApp {
	db := dbm.NewMemDB()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	moduleManager := module.NewManagerFromMap(modules)
	basicModuleManager := module.NewBasicManagerFromManager(moduleManager, nil)
	basicModuleManager.RegisterInterfaces(interfaceRegistry)

	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), authtx.DefaultSignModes)
	bApp := baseapp.NewBaseApp(appName, logger, db, txConfig.TxDecoder(), baseapp.SetChainID(appName))
	bApp.MountKVStores(keys)

	bApp.SetInitChainer(func(ctx sdk.Context, _ *cmtabcitypes.RequestInitChain) (*cmtabcitypes.ResponseInitChain, error) {
		for _, mod := range modules {
			if m, ok := mod.(module.HasGenesis); ok {
				m.InitGenesis(ctx, appCodec, m.DefaultGenesis(appCodec))
			}
		}

		return &cmtabcitypes.ResponseInitChain{}, nil
	})

	bApp.SetBeginBlocker(func(_ sdk.Context) (sdk.BeginBlock, error) {
		return moduleManager.BeginBlock(sdkCtx)
	})
	bApp.SetEndBlocker(func(_ sdk.Context) (sdk.EndBlock, error) {
		return moduleManager.EndBlock(sdkCtx)
	})

	router.SetInterfaceRegistry(interfaceRegistry)
	configurator := module.NewConfigurator(appCodec, router, bApp.GRPCQueryRouter())
	err := moduleManager.RegisterServices(configurator)
	if err != nil {
		panic(err)
	}
	bApp.SetMsgServiceRouter(router)

	if keys[consensusparamtypes.StoreKey] != nil {
		// set baseApp param store
		consensusParamsKeeper := consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]), authtypes.NewModuleAddress("gov").String(), runtime.EventService{})
		bApp.SetParamStore(consensusParamsKeeper.ParamsStore)

		if err := bApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("failed to load application version from store: %w", err))
		}

		if _, err := bApp.InitChain(&cmtabcitypes.RequestInitChain{ChainId: appName, ConsensusParams: simtestutil.DefaultConsensusParams}); err != nil {
			panic(fmt.Errorf("failed to initialize application: %w", err))
		}
	} else {
		if err := bApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("failed to load application version from store: %w", err))
		}

		if _, err := bApp.InitChain(&cmtabcitypes.RequestInitChain{ChainId: appName}); err != nil {
			panic(fmt.Errorf("failed to initialize application: %w", err))
		}
	}

	_, err = bApp.Commit()
	if err != nil {
		panic(err)
	}

	ctx := sdkCtx.WithBlockHeader(cmtproto.Header{ChainID: appName, Time: time.Now()}).WithIsCheckTx(true)

	return &IntegationApp{
		BaseApp:       bApp,
		logger:        logger,
		ctx:           ctx,
		moduleManager: *moduleManager,
		queryHelper:   baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry),
	}
}

// RunMsg provides the ability to run a message and return the response.
// In order to run a message, the application must have a handler for it.
// These handlers are registered on the application message service router.
// The result of the message execution is returned as an Any type.
// That any type can be unmarshaled to the expected response type.
// If the message execution fails, an error is returned.
func (app *IntegationApp) RunMsg(msg sdk.Msg, option ...integration.Option) (*codectypes.Any, error) {
	// set options
	cfg := &integration.Config{}
	for _, opt := range option {
		opt(cfg)
	}

	if cfg.AutomaticCommit {
		defer func() {
			if _, err := app.Commit(); err != nil {
				fmt.Println("error while committing:", err)
			}
		}()
	}

	if cfg.AutomaticFinalizeBlock {
		height := app.LastBlockHeight() + 1
		if _, err := app.FinalizeBlock(&cmtabcitypes.RequestFinalizeBlock{Height: height}); err != nil {
			return nil, fmt.Errorf("failed to run finalize block: %w", err)
		}
	}

	app.logger.Info("Running msg", "msg", msg.String())

	handler := app.MsgServiceRouter().Handler(msg)
	if handler == nil {
		return nil, fmt.Errorf("handler is nil, can't route message %s: %+v", sdk.MsgTypeURL(msg), msg)
	}

	msgResult, err := handler(app.ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to execute message %s: %w", sdk.MsgTypeURL(msg), err)
	}

	var response *codectypes.Any
	if len(msgResult.MsgResponses) > 0 {
		msgResponse := msgResult.MsgResponses[0]
		if msgResponse == nil {
			return nil, fmt.Errorf("got nil msg response %s in message result: %s", sdk.MsgTypeURL(msg), msgResult.String())
		}

		response = msgResponse
	}

	return response, nil
}

// Context returns the application context. It can be unwrapped to a sdk.Context,
// with the sdk.UnwrapSDKContext function.
func (app *IntegationApp) Context() sdk.Context {
	return app.ctx
}

// AddTime adds time to the application context.
func (app *IntegationApp) AddTime(seconds int64) {
	newTime := app.ctx.BlockHeader().Time.Add(time.Duration(int64(time.Second) * seconds))
	app.ctx = app.ctx.WithBlockTime(newTime)
}

// AddBlock increments the block number of the application context.
func (app *IntegationApp) AddBlock() {
	app.ctx = app.ctx.WithBlockHeight(app.ctx.BlockHeader().Height + 1)
}

// QueryHelper returns the application query helper.
// It can be used when registering query services.
func (app *IntegationApp) QueryHelper() *baseapp.QueryServiceTestHelper {
	return app.queryHelper
}

// CreateMultiStore is a helper for setting up multiple stores for provided modules.
func CreateMultiStore(keys map[string]*storetypes.KVStoreKey, logger log.Logger) storetypes.CommitMultiStore {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())

	for key := range keys {
		cms.MountStoreWithDB(keys[key], storetypes.StoreTypeIAVL, db)
	}

	_ = cms.LoadLatestVersion()
	return cms
}
