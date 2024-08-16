package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

const (
	// FlagKeyFile defines a flag to add arbitrary data to a data proxy.
	FlagMemo = "memo"
)

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		RegisterDataProxy(),
	)
	return cmd
}

// AddKey returns the command for adding a new key and uploading its
// public key on chain at a given index.
func RegisterDataProxy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [payout_address] [fee] [public_key_hex] [signature_hex] --from [admin_address]",
		Short: "Register a data proxy using a signed message generated with the data-proxy cli",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fee, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			memo, _ := cmd.Flags().GetString(FlagMemo)

			msg := &types.MsgRegisterDataProxy{
				AdminAddress:  clientCtx.GetFromAddress().String(),
				PayoutAddress: args[0],
				Fee:           &fee,
				PubKey:        args[2],
				Signature:     args[3],
				Memo:          memo,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagMemo, "", "Optionally add a description to the data proxy config")
	return cmd
}

// TODO Edit tx
