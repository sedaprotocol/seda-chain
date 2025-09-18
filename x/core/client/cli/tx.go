package cli

import (
	"encoding/base64"
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
		AcceptOwnership(),
		TransferOwnership(),
		AddToAllowlist(),
		RemoveFromAllowlist(),
		Pause(),
		Unpause(),
	)
	return cmd
}

func AcceptOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept-ownership --from [pending_owner_address]",
		Short: "Accept ownership of the core module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgAcceptOwnership{
				Sender: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func TransferOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-ownership [new_owner_address] --from [owner_address]",
		Short: "Transfer ownership of the core module to a new owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgTransferOwnership{
				Sender:   clientCtx.GetFromAddress().String(),
				NewOwner: args[0],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func AddToAllowlist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-to-allowlist [public_key] --from [owner_address]",
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

func RemoveFromAllowlist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-from-allowlist [public_key] --from [owner_address]",
		Short: "Remove an executor identity from the allowlist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, err := hex.DecodeString(args[0]); err != nil {
				return fmt.Errorf("public key must be a valid hex string")
			}

			msg := &types.MsgRemoveFromAllowlist{
				Sender:    clientCtx.GetFromAddress().String(),
				PublicKey: args[0],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func Pause() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause --from [owner_address]",
		Short: "Pause the core module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgPause{
				Sender: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func Unpause() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpause --from [owner_address]",
		Short: "Unpause the core module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUnpause{
				Sender: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func PostDataRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-data-request",
		Short: "Post a data request to the core module",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, err := hex.DecodeString(args[0]); err != nil {
				return fmt.Errorf("public key must be a valid hex string")
			}

			msg := &types.MsgPostDataRequest{
				Sender: clientCtx.GetFromAddress().String(),
			}

			if rf, _ := cmd.Flags().GetUint32("replication-factor"); rf != 0 {
				msg.ReplicationFactor = rf
			}

			if version, _ := cmd.Flags().GetString("version"); version != "" {
				msg.Version = version
			}

			if memo, _ := cmd.Flags().GetString("memo"); memo != "" {
				msg.Memo = []byte(base64.StdEncoding.EncodeToString([]byte(memo)))
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Uint32P("replication-factor", "rf", 1, "replication factor for the data request")
	cmd.Flags().StringP("version", "v", "v1.0.0", "version for the data request")
	cmd.Flags().StringP("memo", "m", "", "optional memo for the data request")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
