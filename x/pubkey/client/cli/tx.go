package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
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

// AddKey returns the command for generating the SEDA keys and
// uploading their public keys on chain.
func AddKey(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-seda-keys",
		Short: "Generate the SEDA keys and upload their public keys.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			serverCfg := server.GetServerContextFromCmd(cmd).Config

			fromAddr := clientCtx.GetFromAddress()
			if fromAddr.Empty() {
				return fmt.Errorf("set the from address using --from flag")
			}
			valAddr, err := ac.BytesToString(fromAddr)
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
					[]privKeyGenerator{secp256k1GenPrivKey},
					filepath.Dir(serverCfg.PrivValidatorKeyFile()),
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

type IndexedPrivKey struct {
	Index   uint32            `json:"index"`
	PrivKey cmtcrypto.PrivKey `json:"priv_key"`
}

// loadSEDAPubKeys loads the SEDA key file from the given path and
// returns a list of index-public key pairs.
func loadSEDAPubKeys(loadPath string) ([]types.IndexedPubKey, error) {
	keysJSONBytes, err := os.ReadFile(loadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SEDA keys from %v: %v", loadPath, err)
	}
	var keys []IndexedPrivKey
	err = cmtjson.Unmarshal(keysJSONBytes, &keys)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SEDA keys from %v: %v", loadPath, err)
	}

	result := make([]types.IndexedPubKey, len(keys))
	for i, key := range keys {
		// Convert to SDK type for app-level use.
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(key.PrivKey.PubKey())
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, err
		}
		result[i] = types.IndexedPubKey{
			Index:  key.Index,
			PubKey: pkAny,
		}
	}
	return result, nil
}

// saveSEDAKeys saves a given list of IndexedPrivKey in the directory
// at dirPath.
func saveSEDAKeys(keys []IndexedPrivKey, dirPath string) error {
	savePath := filepath.Join(dirPath, SEDAKeyFileName)
	if cmtos.FileExists(savePath) {
		return fmt.Errorf("SEDA key file already exists at %s", savePath)
	}
	err := cmtos.EnsureDir(filepath.Dir(savePath), 0o700)
	if err != nil {
		return err
	}
	jsonBytes, err := cmtjson.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SEDA keys: %v", err)
	}
	err = os.WriteFile(savePath, jsonBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write SEDA key file: %v", err)
	}
	return nil
}

type privKeyGenerator func() cmtcrypto.PrivKey

func secp256k1GenPrivKey() cmtcrypto.PrivKey {
	return secp256k1.GenPrivKey()
}

// generateSEDAKeys generates SEDA keys given a list of private key
// generators, saves them to the SEDA key file, and returns the resulting
// index-public key pairs. Index is assigned incrementally in the order
// of the given private key generators. The key file is stored in the
// directory given by dirPath.
func generateSEDAKeys(generators []privKeyGenerator, dirPath string) ([]types.IndexedPubKey, error) {
	keys := make([]IndexedPrivKey, len(generators))
	result := make([]types.IndexedPubKey, len(generators))
	for i, generator := range generators {
		privKey := generator()
		keys[i] = IndexedPrivKey{
			Index:   uint32(i),
			PrivKey: privKey,
		}

		// Convert to SDK type for app-level use.
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(privKey.PubKey())
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		if err != nil {
			return nil, err
		}
		result[i] = types.IndexedPubKey{
			Index:  uint32(i),
			PubKey: pkAny,
		}
	}

	// The key file is placed in the same directory as the validator key file.
	err := saveSEDAKeys(keys, dirPath)
	if err != nil {
		return nil, err
	}
	return result, nil
}
