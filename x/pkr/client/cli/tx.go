package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

const (
	SEDAKeyFileName = "seda_keys.json"

	// FlagKeyFile defines a flag to specify an existing key file.
	FlagKeyFile = "key-file"
)

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd(valAddrCodec address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		AddKey(valAddrCodec),
	)
	return cmd
}

// AddKey returns the command for adding a new key and uploading its
// public key on chain at a given index.
func AddKey(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-key",
		Short: "Generate the SEDA keys and upload their public keys on chain at a given index",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			serverCfg := server.GetServerContextFromCmd(cmd).Config

			valAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			var pks []types.IndexedPubKey
			keyFile, _ := cmd.Flags().GetString(FlagKeyFile)
			if keyFile != "" {
				pks, err = loadSEDAPubKeys(keyFile)
				if err != nil {
					return err
				}
			} else {
				pks, err = generateSEDAKeys(
					serverCfg,
					[]privKeyGenerator{secp256k1GenPrivKey},
				)
				if err != nil {
					return err
				}
			}

			msg := &types.MsgAddKey{
				ValidatorAddr:  valAddr,
				IndexedPubKeys: pks,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagKeyFile, "", "path to an existing SEDA key file")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

type IndexKey struct {
	Index   uint32         `json:"index"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

// saveSEDAKeys saves a given list of IndexKeys at a given path.
func saveSEDAKeys(keys []IndexKey, savePath string) error {
	jsonBytes, err := cmtjson.MarshalIndent(keys, "", "  ") // TODO use simple json.Marshal?
	if err != nil {
		return fmt.Errorf("failed to marshal SEDA keys: %v", err)
	}
	err = os.WriteFile(savePath, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write SEDA key file: %v", err)
	}
	return nil
}

// loadSEDAPubKeys loads the SEDA key file from the given path and
// returns a list of index-public key pairs.
func loadSEDAPubKeys(loadPath string) ([]types.IndexedPubKey, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keys []IndexKey
	err = cmtjson.Unmarshal(keysJSONBytes, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}

	var result []types.IndexedPubKey
	for _, key := range keys {
		pkAny, err := codectypes.NewAnyWithValue(key.PubKey)
		if err != nil {
			return nil, err
		}
		result = append(result, types.IndexedPubKey{
			Index:  key.Index,
			PubKey: pkAny,
		})
	}
	return result, nil
}

type privKeyGenerator func() crypto.PrivKey

func secp256k1GenPrivKey() crypto.PrivKey {
	return secp256k1.GenPrivKey()
}

// generateSEDAKeys generates SEDA keys given a list of private key
// generators, saves them to the SEDA key file, and returns the resulting
// index-public key pairs. Index is assigned incrementally in the order
// of the given private key generators.
func generateSEDAKeys(config *cfg.Config, generators []privKeyGenerator) ([]types.IndexedPubKey, error) {
	var keys []IndexKey
	var result []types.IndexedPubKey
	for i, generator := range generators {
		privKey := generator()
		keys = append(keys, IndexKey{
			Index:   uint32(i),
			PrivKey: privKey,
			PubKey:  privKey.PubKey(),
		})

		pkAny, err := codectypes.NewAnyWithValue(privKey.PubKey())
		if err != nil {
			return nil, err
		}
		result = append(result, types.IndexedPubKey{
			Index:  uint32(i),
			PubKey: pkAny,
		})
	}

	// The key file is placed in the same directory as the validator key file.
	pvKeyFile := config.PrivValidatorKeyFile()
	savePath := filepath.Join(filepath.Dir(pvKeyFile), SEDAKeyFileName)
	if cmtos.FileExists(savePath) {
		return nil, fmt.Errorf("SEDA key file already exists at %s", savePath)
	}
	err := cmtos.EnsureDir(filepath.Dir(pvKeyFile), 0o700)
	if err != nil {
		return nil, err
	}
	saveSEDAKeys(keys, savePath)

	return result, nil
}
