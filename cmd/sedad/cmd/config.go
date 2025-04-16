package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tmcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/app/utils"
)

// AppConfig defines the application configurations. It extends the default
// Cosmos SDK server config with custom SEDA configurations.
type AppConfig struct {
	serverconfig.Config
	SEDAConfig utils.SEDAConfig `mapstructure:"seda_config"`
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// The following code snippet is just for reference.

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the min gas prices to 0.
	srvCfg.MinGasPrices = params.MinimumGasPrice.String()
	srvCfg.Mempool.MaxTxs = params.DefaultMempoolMaxTxs

	// GRPC settings
	srvCfg.GRPC.Enable = true
	srvCfg.GRPC.Address = "0.0.0.0:9090"

	// GRPC Web Settings
	srvCfg.GRPCWeb.Enable = true

	// API Settings
	srvCfg.API.Enable = true
	srvCfg.API.Address = "tcp://0.0.0.0:1317"
	srvCfg.API.EnableUnsafeCORS = true

	config := AppConfig{
		Config:     *srvCfg,
		SEDAConfig: utils.DefaultSEDAConfig(),
	}
	template := serverconfig.DefaultConfigTemplate + utils.DefaultSEDATemplate

	return template, config
}

// initTendermintConfig helps to override default Tendermint Config values.
// return tmcfg.DefaultConfig if no custom configuration is required for the application.
func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()

	// Log Settings
	cfg.LogFormat = "json"
	// cfg.LogLevel

	// RPC Settings
	cfg.RPC.ListenAddress = "tcp://0.0.0.0:26657"

	// Consensus Settings
	cfg.Consensus.TimeoutPropose = time.Duration(7.5 * float64(time.Second))
	cfg.Consensus.TimeoutProposeDelta = time.Duration(0)
	cfg.Consensus.TimeoutCommit = time.Duration(7.5 * float64(time.Second))

	return cfg
}

func preUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pre-upgrade",
		Short: "Pre-upgrade command",
		Long:  "Pre-upgrade command to migrate app.toml for v1.0.0 upgrade",
		Run: func(cmd *cobra.Command, args []string) {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)
			err := migrateAppConfig(serverCtx, clientCtx.HomeDir)
			if err != nil {
				os.Exit(30)
			}
			os.Exit(0)
		},
	}

	return cmd
}

func migrateAppConfig(serverCtx *server.Context, rootDir string) error {
	configPath := filepath.Join(rootDir, "config")
	appConfigPath := filepath.Join(configPath, "app.toml")

	serverconfig.SetConfigTemplate(serverconfig.DefaultConfigTemplate)
	oldConfig := serverconfig.DefaultConfig()
	err := serverCtx.Viper.Unmarshal(oldConfig)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", appConfigPath, err)
	}

	newConfig := AppConfig{
		Config:     *oldConfig,
		SEDAConfig: utils.DefaultSEDAConfig(),
	}
	serverconfig.SetConfigTemplate(serverconfig.DefaultConfigTemplate + utils.DefaultSEDATemplate)
	serverconfig.WriteConfigFile(appConfigPath, newConfig)
	return nil
}
