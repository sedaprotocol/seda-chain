package cli

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/go-bip39"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

const (
	// FlagKeyFile defines a flag to specify an existing key file.
	FlagKeyFile = "key-file"
	// FlagMnemonic defines a flag to generate a key from a mnemonic.
	FlagMnemonic = "mnemonic"
	// FlagNonDeterministic defines a flag to generate a non-deterministic
	// key.
	FlagNonDeterministic = "non-deterministic"
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
			}

			index, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return errorsmod.Wrap(fmt.Errorf("invalid index: %d", index), "invalid index")
			}
			var pk crypto.PubKey
			switch index {
			case 0:
				// VRF key derived using secp256k1
				pk, err = utils.LoadOrGenVRFKey(serverCtx.Config, keyFile, mnemonic)
				if err != nil {
					return errorsmod.Wrap(err, "failed to initialize a new key")
				}
			default:
				panic("unsupported index")
			}

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
