package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

const (
	// FlagKeyFile defines a flag to specify an existing key file.
	FlagKeyFile = "key-file"
	// FlagForceKeyFile defines a flag to specify that the key file should be overwritten if it already exists.
	FlagForceKeyFile = "key-file-force"
	// FlagEncryptionKey defines a flag to specify an existing encryption key.
	FlagEncryptionKey = "key-file-custom-encryption-key"
	// FlagNoEncryption defines a flag to specify that the generated key file should not be encrypted.
	FlagNoEncryption = "key-file-no-encryption"
)

func AddSedaKeysFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(FlagKeyFile, "", "path to an existing SEDA key file")
	cmd.Flags().Bool(FlagForceKeyFile, false, "overwrite the existing key file if it already exists")
	cmd.Flags().Bool(FlagEncryptionKey, false, "use a custom AES encryption key for the SEDA key file (if not set, a random key will be generated)")
	cmd.Flags().Bool(FlagNoEncryption, false, "do not encrypt the generated SEDA key file (if the key file is not provided)")
}

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr := sdk.ValAddress(clientCtx.GetFromAddress())
			if valAddr.Empty() {
				return fmt.Errorf("set the from address using --from flag")
			}
			valStr, err := ac.BytesToString(valAddr)
			if err != nil {
				return err
			}

			pks, err := LoadOrGenerateSEDAKeys(cmd, valAddr)
			if err != nil {
				return err
			}

			msg := &types.MsgAddKey{
				ValidatorAddr:  valStr,
				IndexedPubKeys: pks,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	AddSedaKeysFlagsToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
