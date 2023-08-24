package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/cli"
	"github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/go-bip39"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/cmd/seda-chaind/utils"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"
	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"

	// Default things
	// BondDenom = "seda"
	ChainID = "sedachain"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))

	return err
}

func validateOrGenerateMnemonic(recover bool, cmd *cobra.Command) (string, error) {
	var mnemonic string
	if recover {
		inBuf := bufio.NewReader(cmd.InOrStdin())
		value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
		if err != nil {
			return "", err
		}

		mnemonic = value
		if !bip39.IsMnemonicValid(mnemonic) {
			return "", errors.New("invalid mnemonic")
		}
	}

	return mnemonic, nil
}

// preserve old logic for if we want to create a new network
// though its slightly modified to set default settings.
func newNetworkCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			recover, _ := cmd.Flags().GetBool(FlagRecover)
			mnemonic, err := validateOrGenerateMnemonic(recover, cmd)
			if err != nil {
				return err
			}

			// Get initial height
			initHeight, _ := cmd.Flags().GetInt64(flags.FlagInitHeight)
			if initHeight < 1 {
				initHeight = 1
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

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

			genDoc := &types.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = types.GenesisDocFromFile(genFile)
				if err != nil {
					return errors.Wrap(err, "Failed to read genesis doc from file")
				}
			}

			genDoc.ChainID = ChainID
			genDoc.Validators = nil
			genDoc.AppState = appState
			genDoc.InitialHeight = initHeight

			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return errors.Wrap(err, "Failed to export genesis file")
			}
			toPrint := newPrintInfo(config.Moniker, ChainID, nodeID, "", appState)
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Int64(flags.FlagInitHeight, 1, "specify the initial block height at genesis")

	return cmd
}

func joinNetwork(network, configDir, genesisFilePath, mnemonic string, config *cfg.Config) error {
	err := utils.DownloadGitFiles(network, configDir)
	if err != nil {
		return errors.Wrapf(err, "failed to download network `%s` genesis files", network)
	}

	bytes, err := os.ReadFile(genesisFilePath)
	if err != nil {
		return err
	}

	var genesisExistingState map[string]json.RawMessage
	err = json.Unmarshal(bytes, &genesisExistingState)
	if err != nil {
		return err
	}

	genesisState, err := json.MarshalIndent(genesisExistingState, "", " ")
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal network `%s` genesis state", network)
	}

	nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
	if err != nil {
		return err
	}

	toPrint := newPrintInfo(config.Moniker, ChainID, nodeID, "", genesisState)

	return displayInfo(toPrint)
}

func existingNetworkComand(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network [moniker]",
		Short: "Grabs an existing network genesis configuration.",
		Long:  `Initialize validators's and node's configuration files from an existing configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			recover, _ := cmd.Flags().GetBool(FlagRecover)
			mnemonic, err := validateOrGenerateMnemonic(recover, cmd)
			if err != nil {
				return err
			}

			// get the value of the network flag
			network, _ := cmd.Flags().GetString("network")
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)
			configDir := filepath.Join(config.RootDir, "config")
			genesisFilePath := filepath.Join(configDir, "genesis.json")
			// use os.Stat to check if the file exists
			_, err = os.Stat(genesisFilePath)
			if !overwrite && !os.IsNotExist(err) {
				return fmt.Errorf("genesis.json file already exists: %v", genesisFilePath)
			}

			// If we are overwriting the genesis make sure to remove gentx folder
			// this is in case they are switching to a different network
			if overwrite {
				gentxDir := filepath.Join(configDir, "gentx")
				err = os.RemoveAll(gentxDir)
				if err != nil {
					return err
				}
			}

			// TODO should turn the insides here into a function for when we have more than one network
			switch network {
			case "devnet":
				return joinNetwork(network, configDir, genesisFilePath, mnemonic, config)
			default:
				return fmt.Errorf("unsupported network type: %s", network)
			}
		},
	}

	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().StringP("network", "n", "devnet", "Specify the type of network to initialize (e.g., 'mainnet', 'testnet', 'devnet')")

	return cmd
}

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <new | newtwork> [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
	}

	cmd.AddCommand(newNetworkCmd(mbm, defaultNodeHome))
	cmd.AddCommand(existingNetworkComand(mbm, defaultNodeHome))

	return cmd
}
