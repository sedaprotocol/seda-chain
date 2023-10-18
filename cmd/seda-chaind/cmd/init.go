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
	cmtos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client/input"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	"github.com/sedaprotocol/seda-chain/cmd/seda-chaind/utils"
)

type printInfo struct {
	Moniker string `json:"moniker" yaml:"moniker"`
	ChainID string `json:"chain_id" yaml:"chain_id"`
	NodeID  string `json:"node_id" yaml:"node_id"`
	Seeds   string `json:"seeds" yaml:"seeds"`
}

func newPrintInfo(moniker, chainID, nodeID, seeds string) printInfo {
	return printInfo{
		Moniker: moniker,
		ChainID: chainID,
		NodeID:  nodeID,
		Seeds:   seeds,
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

// If validator key file exists, create and save an empty validator state file.
func configureValidatorFiles(config *cfg.Config) error {
	keyFilePath := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0o777); err != nil {
		return fmt.Errorf("could not create directory %q: %w", filepath.Dir(keyFilePath), err)
	}
	stateFilePath := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0o777); err != nil {
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
	err = utils.DownloadGitFiles(network, configDir)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to download files for network `%s`", network)
	}

	// check genesis file
	genFile := config.GenesisFile()
	jsonBlob, err := os.ReadFile(genFile)
	if err != nil {
		return "", "", err
	}
	genDoc, err := types.GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return "", "", errors.Wrapf(err, "error reading GenesisDoc at %s", genFile)
	}
	chainID = genDoc.ChainID

	// obtain seeds from seeds file, if exists, and write to config file
	seedsBytes, err := os.ReadFile(seedsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", "", err
		}
	}
	seeds = string(seedsBytes)

	config.P2P.Seeds = seeds
	config.Moniker = moniker
	cfg.WriteConfigFile(filepath.Join(configDir, "config.toml"), config)

	return chainID, seeds, nil
}
