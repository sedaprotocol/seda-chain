package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pubkey/types"
)

const (
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
				pks, err = utils.LoadSEDAPubKeys(keyFile)
				if err != nil {
					return err
				}
			} else {
				pks, err = utils.GenerateSEDAKeys(filepath.Dir(serverCfg.PrivValidatorKeyFile()))
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
