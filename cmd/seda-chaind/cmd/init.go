package cmd

import (
	"fmt"
	"seda-chain/app"

	"time"

	tmcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
)

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	// get the original command
	originalCmd := genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome)

	// store the original RunE function
	originalRunE := originalCmd.RunE

	// wrap the RunE function to add additional behavior
	originalCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// call the original RunE function
		err := originalRunE(cmd, args)
		if err != nil {
			return err
		}

		// TODO how to tell if initing a node for testnet vs mainnet vs etc.
		fmt.Println("Changed the init command")

		return nil
	}

	return originalCmd
}

// initTendermintConfig helps to override default Tendermint Config values.
// return tmcfg.DefaultConfig if no custom configuration is required for the application.
func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()

	// Log Settings
	cfg.LogFormat = "json"
	// TODO how to tell if initing a node for testnet vs mainnet vs etc.
	// cfg.LogLevel

	// RPC Settings
	cfg.RPC.GRPCListenAddress = "0.0.0.0:26657"

	// Consensus Settings
	cfg.Consensus.TimeoutPropose = time.Duration(7.5 * float64(time.Second))
	cfg.Consensus.TimeoutProposeDelta = time.Duration(0)
	cfg.Consensus.TimeoutCommit = time.Duration(7.5 * float64(time.Second))

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// The following code snippet is just for reference.

	type CustomAppConfig struct {
		serverconfig.Config
	}

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
	srvCfg.MinGasPrices = "0seda"

	// GRPC settings
	srvCfg.GRPC.Enable = true
	srvCfg.GRPC.Address = "0.0.0.0:9090"

	// GRPC Web Settings
	srvCfg.GRPCWeb.Enable = true
	srvCfg.GRPCWeb.Address = "0.0.0.0:9091"
	srvCfg.GRPCWeb.EnableUnsafeCORS = true

	// API Settings
	srvCfg.API.Enable = true
	srvCfg.API.Address = "tcp://0.0.0.0:1317"
	srvCfg.API.EnableUnsafeCORS = true

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
	}
	customAppTemplate := serverconfig.DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}
