package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	"github.com/sedaprotocol/seda-chain/app/utils"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"
	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"
	// FlagNetwork defines a flag to indicate which network to connect to.
	FlagNetwork = "network"
)

func JoinNetworkCommand(_ module.BasicManager, _ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join [moniker]",
		Short: "Grabs an existing network configuration and initializes node based on it",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Initialize validator and node configuration files for an existing network.

Example:
$ %s join moniker --network devnet
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
			if recoverFlag, _ := cmd.Flags().GetBool(FlagRecover); recoverFlag {
				mnemonic, err = readInMnemonic(cmd)
				if err != nil {
					return err
				}
			}

			network, _ := cmd.Flags().GetString(FlagNetwork)
			var seeds, chainID string
			switch network {
			case "mainnet", "devnet", "testnet", "localnet":
				chainID, seeds, err = downloadAndApplyNetworkConfig(network, args[0], config)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported network type: %s", network)
			}

			// TO-DO remove (See: https://github.com/sedaprotocol/seda-chain/pull/76#issuecomment-1762303200)
			// If validator key file exists, create and save an empty validator state file.
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
			toPrint := utils.NewPrintInfo(config.Moniker, chainID, nodeID, seeds)
			return utils.DisplayInfo(toPrint)
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
