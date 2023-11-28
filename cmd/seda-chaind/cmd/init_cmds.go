package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/sedaprotocol/seda-chain/app/params"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"
	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"
	// FlagNetwork defines a flag to indicate which network to connect to.
	FlagNetwork = "network"

	// Default things
	defaultChainID = "sedachain"
)

// add initialization commands
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
	}
	cmd.AddCommand(newNetworkCmd(mbm, defaultNodeHome))
	cmd.AddCommand(joinNetworkCommand(mbm, defaultNodeHome))
	return cmd
}

// preserve old logic for if we want to create a new network
// though its slightly modified to set default settings.
func newNetworkCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validator and node configuration files for a new network.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			var mnemonic string
			var err error
			if recover, _ := cmd.Flags().GetBool(FlagRecover); recover {
				mnemonic, err = readInMnemonic(cmd)
				if err != nil {
					return err
				}
			}

			// get chain ID
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)

			// get initial height
			initHeight, _ := cmd.Flags().GetInt64(flags.FlagInitHeight)
			if initHeight < 1 {
				initHeight = 1
			}

			// initialize node
			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			// write to config file and genesis file and display info
			config.Moniker = args[0]

			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)
			genFile := config.GenesisFile()
			// use os.Stat to check if the file exists
			_, err = os.Stat(genFile)
			if !overwrite && !os.IsNotExist(err) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			sdk.DefaultBondDenom = params.DefaultBondDenom
			appGenState := mbm.DefaultGenesis(cdc)

			appState, err := json.MarshalIndent(appGenState, "", " ")
			if err != nil {
				return errors.Wrap(err, "Failed to marshal default genesis state")
			}

			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			}

			appGenesis := &genutiltypes.AppGenesis{
				ChainID:  chainID,
				AppState: appState,
				Consensus: &genutiltypes.ConsensusGenesis{
					Validators: nil,
				},
				InitialHeight: initHeight,
			}

			if err = genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
				return errors.Wrap(err, "Failed to export genesis file")
			}
			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "")
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, defaultChainID, "genesis file chain-id")
	cmd.Flags().Int64(flags.FlagInitHeight, 1, "specify the initial block height at genesis")

	return cmd
}

func joinNetworkCommand(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join [moniker]",
		Short: "Grabs an existing network configuration and initializes node based on it",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Initialize validator and node configuration files for an existing network.

Example:
$ %s init join moniker --network devnet
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			var mnemonic string
			var err error
			if recover, _ := cmd.Flags().GetBool(FlagRecover); recover {
				mnemonic, err = readInMnemonic(cmd)
				if err != nil {
					return err
				}
			}

			network, _ := cmd.Flags().GetString(FlagNetwork)
			var seeds, chainID string
			if network == "mainnet" || network == "devnet" || network == "testnet" || network == "localnet" {
				chainID, seeds, err = downloadAndApplyNetworkConfig(network, args[0], config)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unsupported network type: %s", network)
			}

			// configure validator files
			err = configureValidatorFiles(config)
			if err != nil {
				return err
			}

			// initialize the node
			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			// genesis and config files already written - display info
			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, seeds)
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().StringP(FlagNetwork, "n", "devnet", "specify the name of network to initialize (e.g., 'mainnet', 'testnet', 'devnet', 'localnet')")
	err := cmd.MarkFlagRequired(FlagNetwork)
	if err != nil {
		panic(err)
	}

	return cmd
}
