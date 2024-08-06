package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	cfg "github.com/cometbft/cometbft/config"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtos "github.com/cometbft/cometbft/libs/os"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

const (
	flagHDPath    = "hd-path"
	flagAddrIndex = "addr-index"
	flagCoinType  = "coin-type"
	flagAccount   = "account"

	// FlagKeyFile defines a flag to specify an existing key file.
	FlagKeyFile = "key-file"
	// FlagMnemonic defines a flag to generate a key from a mnemonic.
	FlagMnemonic = "mnemonic"
	// FlagNonDeterministic defines a flag to generate a non-deterministic
	// key.
	FlagNonDeterministic = "non-deterministic"

	mnemonicEntropySize = 256
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
		Use:   "add-key [index]",
		Short: "Generate a key and upload its public key on chain at a given index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			serverCtx := server.GetServerContextFromCmd(cmd)

			valAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			isNonDet, _ := cmd.Flags().GetBool(FlagNonDeterministic)
			isMnemonic, _ := cmd.Flags().GetBool(FlagMnemonic)
			keyFile, _ := cmd.Flags().GetString(FlagKeyFile)
			var isKeyFile bool
			if keyFile != "" {
				isKeyFile = true
			}
			if ok := isOnlyOneTrue(isMnemonic, isKeyFile, isNonDet); !ok {
				return fmt.Errorf("set one of the flags: %s, %s, or %s", FlagMnemonic, FlagKeyFile, FlagNonDeterministic)
			}

			// mnemonic & bip39 passphrase
			var bip39Passphrase string
			var mnemonic string
			if isMnemonic {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			} else {
				// read entropy seed straight from cmtcrypto.Rand and convert to mnemonic
				entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
				if err != nil {
					return err
				}

				mnemonic, err = bip39.NewMnemonic(entropySeed)
				if err != nil {
					return err
				}
			}

			// Index & algo & name
			index, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return errorsmod.Wrap(fmt.Errorf("invalid index: %d", index), "invalid index")
			}
			var name string
			var algo keyring.SignatureAlgo
			switch index {
			case 0:
				name = "vrf_key.json" // TODO without .json
				algo = hd.Secp256k1
			default:
				panic("unsupported index")
			}

			// HD Path
			coinType, _ := cmd.Flags().GetUint32(flagCoinType)
			account, _ := cmd.Flags().GetUint32(flagAccount)
			addrIndex, _ := cmd.Flags().GetUint32(flagAddrIndex)
			hdPath, _ := cmd.Flags().GetString(flagHDPath)
			if len(hdPath) == 0 {
				hdPath = hd.CreateHDPath(coinType, account, addrIndex).String()
			}

			// Derive and save key.
			pk, err := deriveKeyAndSaveToFile(serverCtx.Config, name, mnemonic, bip39Passphrase, hdPath, algo)
			if err != nil {
				return err
			}

			// Generate and broadcast tx.
			pkAny, err := codectypes.NewAnyWithValue(pk)
			if err != nil {
				return err
			}
			msg := &types.MsgAddKey{
				ValidatorAddr: valAddr,
				Index:         uint32(index),
				PubKey:        pkAny,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagHDPath, "", "Manual HD Path derivation (overrides BIP44 config)")
	cmd.Flags().Uint32(flagCoinType, sdk.GetConfig().GetCoinType(), "coin type number for HD derivation")
	cmd.Flags().Uint32(flagAccount, 0, "Account number for HD derivation (less than equal 2147483647)")
	cmd.Flags().Uint32(flagAddrIndex, 0, "Address index number for HD derivation (less than equal 2147483647)")
	cmd.Flags().Bool(FlagNonDeterministic, false, "generate a key non-deterministically")
	cmd.Flags().String(FlagKeyFile, "", "path to an existing key file")
	cmd.Flags().Bool(FlagMnemonic, false, "provide master seed from which the new key is derived")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// isOnlyOneTrue returns true if only one of the boolean variables is
// true.
func isOnlyOneTrue(bools ...bool) bool {
	trueCount := 0
	for _, b := range bools {
		if b {
			trueCount++
		}
	}
	return trueCount == 1
}

func deriveKeyAndSaveToFile(config *cfg.Config, keyFileName, mnemonic, bip39Passphrase, hdPath string, algo keyring.SignatureAlgo) (pubKey crypto.PubKey, err error) {
	derivedPriv, err := algo.Derive()(mnemonic, bip39Passphrase, hdPath)
	if err != nil {
		return nil, err
	}
	privKey := algo.Generate()(derivedPriv)

	// The key file is placed in the same directory as the validator key file.
	pvKeyFile := config.PrivValidatorKeyFile()
	savePath := filepath.Join(filepath.Dir(pvKeyFile), keyFileName)
	if cmtos.FileExists(savePath) {
		return nil, fmt.Errorf("key file already exists at %s", savePath)
	}
	err = cmtos.EnsureDir(filepath.Dir(pvKeyFile), 0o700)
	if err != nil {
		return nil, err
	}

	keyFile := struct {
		PrivKey crypto.PrivKey `json:"priv_key"`
		PubKey  crypto.PubKey  `json:"pub_key"`
	}{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
	}

	jsonBytes, err := cmtjson.MarshalIndent(keyFile, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %v", err)
	}

	err = os.WriteFile(savePath, jsonBytes, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to write key file: %v", err)
	}

	return privKey.PubKey(), nil
}
