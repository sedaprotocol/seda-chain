package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoregistry"

	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	tmcli "github.com/cometbft/cometbft/libs/cli"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/client/v2/autocli/flag"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/sedaprotocol/seda-chain/app"
	appparams "github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
	_ "github.com/sedaprotocol/seda-chain/client/docs/statik" // for swagger docs
	"github.com/sedaprotocol/seda-chain/cmd/sedad/gentx"
)

const defaultWasmDir = "wasm"

// NewRootCmd creates a new root command for a Cosmos SDK application
func NewRootCmd() *cobra.Command {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(appparams.Bech32PrefixAccAddr, appparams.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(appparams.Bech32PrefixValAddr, appparams.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(appparams.Bech32PrefixConsAddr, appparams.Bech32PrefixConsPub)
	cfg.SetAddressVerifier(wasmtypes.VerifyAddressLen())
	cfg.Seal()

	// "Pre-instantiate" the application for getting the injected/configured
	// encoding configuration note, this is not necessary when using app wiring,
	// as depinject can be directly used (see root_v2.go)
	tempApp := app.NewApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		simtestutil.NewAppOptionsWithFlagHome(tempDir()),
		tempDir(),
		baseapp.SetChainID("tempchainid"),
	)
	encodingConfig := app.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Marshaler:         tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   "sedad",
		Short: "Start sedachain node",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL. This sign mode
			// is only available if the client is online.
			if !initClientCtx.Offline {
				//nolint:gocritic
				enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)

				txConfigOpts := tx.ConfigOptions{
					EnabledSignModes:           enabledSignModes,
					TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
				}
				txConfig, err := tx.NewTxConfigWithOptions(
					initClientCtx.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			appCfgTemplate, appCfg := initAppConfig()
			cmtCfg := initCometConfig()
			return server.InterceptConfigsPreRunHandler(cmd, appCfgTemplate, appCfg, cmtCfg)
		},
	}

	initRootCmd(rootCmd, encodingConfig, tempApp.BasicModuleManager())
	addRosettaCmd(rootCmd, encodingConfig)

	autoCliOpts := tempApp.AutoCliOpts()
	initClientCtx, _ = config.ReadFromClientConfig(initClientCtx)
	autoCliOpts.ClientCtx = initClientCtx

	builder := &autocli.Builder{
		Builder: flag.Builder{
			TypeResolver:          protoregistry.GlobalTypes,
			FileResolver:          autoCliOpts.ClientCtx.InterfaceRegistry,
			AddressCodec:          autoCliOpts.AddressCodec,
			ValidatorAddressCodec: autoCliOpts.ValidatorAddressCodec,
			ConsensusAddressCodec: autoCliOpts.ConsensusAddressCodec,
		},
		GetClientConn: func(cmd *cobra.Command) (grpc.ClientConnInterface, error) {
			return client.GetClientQueryContext(cmd)
		},
		AddQueryConnFlags: sdkflags.AddQueryFlagsToCmd,
		AddTxConnFlags:    sdkflags.AddTxFlagsToCmd,
	}
	if err := autoCliOpts.EnhanceRootCommandWithBuilder(rootCmd, builder); err != nil {
		panic(err)
	}

	return rootCmd
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig, basicManager module.BasicManager) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	gentxModule := app.ModuleBasics[genutiltypes.ModuleName].(genutil.AppModuleBasic)
	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		JoinNetworkCommand(basicManager, app.DefaultNodeHome),
		gentx.CollectGenTxsCmd(
			banktypes.GenesisBalancesIterator{},
			app.DefaultNodeHome,
			gentxModule.GenTxValidator,
			encodingConfig.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		),
		gentx.GenTxCmd(
			basicManager,
			encodingConfig.TxConfig,
			banktypes.GenesisBalancesIterator{},
			app.DefaultNodeHome,
			encodingConfig.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		),
		genutilcli.ValidateGenesisCmd(basicManager),
		addGenesisAccountCmd(app.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		preUpgradeCmd(),
	)

	// add server commands
	addServerCommands(rootCmd, newApp, app.DefaultNodeHome)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(basicManager),
		txCommand(basicManager),
		keys.Commands(),
	)
}

func addServerCommands(rootCmd *cobra.Command, appCreator servertypes.AppCreator, defaultNodeHome string) {
	cometCmd := &cobra.Command{
		Use:     "comet",
		Aliases: []string{"cometbft", "tendermint"},
		Short:   "CometBFT subcommands",
	}

	cometCmd.AddCommand(
		server.ShowNodeIDCmd(),
		server.ShowValidatorCmd(),
		server.ShowAddressCmd(),
		server.VersionCmd(),
		cmtcmd.ResetAllCmd,
		cmtcmd.ResetStateCmd,
		server.BootstrapStateCmd(appCreator),
	)

	startCmd := server.StartCmdWithOptions(appCreator, defaultNodeHome,
		server.StartCmdOptions{
			AddFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool(crisis.FlagSkipGenesisInvariants, false, "Skip x/crisis invariants check on startup")
			},
		},
	)
	// Overwrite PreRunE of start command to add config key for node start.
	startCmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		serverCtx := server.GetServerContextFromCmd(cmd)

		// Bind flags to the Context's Viper so the app construction can set
		// options accordingly.
		if err := serverCtx.Viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		utils.SetNodeStart(serverCtx)

		_, err := server.GetPruningOptionsFromFlags(serverCtx.Viper)
		return err
	}

	serverCmds := []*cobra.Command{
		startCmd,
		server.ExportCmd(appExport, defaultNodeHome),
		server.NewRollbackCmd(appCreator, defaultNodeHome),
		server.ModuleHashByHeightQuery(appCreator),
	}

	rootCmd.AddCommand(
		append([]*cobra.Command{cometCmd, version.NewVersionCommand()}, serverCmds...)...,
	)
}

// queryCommand returns the sub-command to send queries to the app
func queryCommand(_ module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
		app.GetSEDASignerStatus(),
	)

	return cmd
}

// txCommand returns the sub-command to send transactions to the app
func txCommand(_ module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		sdkflags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	return cmd
}

// newApp creates a new Cosmos SDK app
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	homeDir := cast.ToString(appOpts.Get(sdkflags.FlagHome))
	chainID := cast.ToString(appOpts.Get(sdkflags.FlagChainID))
	if chainID == "" {
		// fallback to genesis chain-id
		appGenesis, err := genutiltypes.AppGenesisFromFile(filepath.Join(homeDir, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}

		chainID = appGenesis.ChainID
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(sdkflags.FlagHome)), "data", "snapshots")
	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
	)

	return app.NewApp(
		logger,
		db,
		traceStore,
		true,
		skipUpgradeHeights,
		cast.ToString(appOpts.Get(sdkflags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		appOpts,
		defaultWasmDir,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetIAVLCacheSize(cast.ToInt(appOpts.Get(server.FlagIAVLCacheSize))),
		baseapp.SetIAVLDisableFastNode(cast.ToBool(appOpts.Get(server.FlagDisableIAVLFastNode))),
		baseapp.SetChainID(chainID),
	)
}

// appExport creates a new simapp (optionally at a given height)
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	homePath, ok := appOpts.Get(sdkflags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	app := app.NewApp(
		logger,
		db,
		traceStore,
		height == -1, // -1: no height provided
		map[int64]bool{},
		homePath,
		uint(1),
		appOpts,
		defaultWasmDir,
	)

	if height != -1 {
		if err := app.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", "tempchain")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	return dir
}
