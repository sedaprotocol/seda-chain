package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/privval"

	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func readInMnemonic(cmd *cobra.Command) (string, error) {
	var mnemonic string
	inBuf := bufio.NewReader(cmd.InOrStdin())
	value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
	if err != nil {
		return "", err
	}

	mnemonic = value
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", errors.New("invalid mnemonic")
	}
	return mnemonic, nil
}

func configureValidatorFiles(config *cfg.Config) error {
	keyFilePath := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0o700); err != nil {
		return fmt.Errorf("could not create directory %q: %w", filepath.Dir(keyFilePath), err)
	}
	stateFilePath := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0o700); err != nil {
		return fmt.Errorf("could not create directory %q: %w", filepath.Dir(stateFilePath), err)
	}

	var pv *privval.FilePV
	if cmtos.FileExists(keyFilePath) {
		pv = privval.LoadFilePVEmptyState(keyFilePath, stateFilePath)
		pv.Save()
	}
	return nil
}

// downloadAndApplyNetworkConfig() downloads network files from seda-networks
// repo. Then it validates the genesis file and writes the seed list and given
// moniker to the config file.
func downloadAndApplyNetworkConfig(network, moniker string, config *cfg.Config) (chainID, seeds string, err error) {
	configDir := filepath.Join(config.RootDir, "config")

	// use os.Stat to check if the file exists
	if cmtos.FileExists(config.GenesisFile()) {
		return "", "", fmt.Errorf("genesis.json file already exists: %v", config.GenesisFile())
	}
	seedsFile := filepath.Join(configDir, "seeds.txt")
	if cmtos.FileExists(seedsFile) {
		return "", "", fmt.Errorf("seeds.txt file already exists: %v", seedsFile)
	}

	// download files from seda-networks repo
	err = DownloadGitFiles(network, configDir)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to download files for network `%s`", network)
	}

	// check genesis file
	appGenesis, err := types.AppGenesisFromFile(config.GenesisFile())
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to read genesis doc file %s", config.GenesisFile())
	}
	chainID = appGenesis.ChainID

	// obtain seeds from seeds file, if exists, and write to config file
	seedsBytes, err := os.ReadFile(seedsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", "", errors.Wrapf(err, "error reading seeds file at %s", seedsFile)
		}
	} else {
		seeds = strings.TrimRight(string(seedsBytes), " \n")
		config.P2P.Seeds = seeds
	}

	config.Moniker = moniker
	cfg.WriteConfigFile(filepath.Join(configDir, "config.toml"), config)

	return chainID, seeds, nil
}
