package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// GetTxCmd returns the CLI transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		AddToAllowlist(),
	)
	return cmd
}

func AddToAllowlist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-to-allowlist [public_key]",
		Short: "Add an executor identity to the allowlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, err := hex.DecodeString(args[0]); err != nil {
				return fmt.Errorf("public key must be a valid hex string")
			}

			msg := &types.MsgAddToAllowlist{
				Sender:    clientCtx.GetFromAddress().String(),
				PublicKey: args[0],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
