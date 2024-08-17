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
	FlagMemo             = "memo"
	FlagNewPayoutAddress = "payout-address"
	FlagNewFee           = "fee"
	FlagFeeUpdateDelay   = "fee-delay"
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
		EditDataProxy(),
	)
	return cmd
}

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

func EditDataProxy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [public_key_hex] --from [admin_address]",
		Short: "Edit an existing data proxy. Payout address and memo take effect instantly, fee updates are scheduled according to the minimum delay or a custom delay",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			newMemo, _ := cmd.Flags().GetString(FlagMemo)
			newPayoutAddress, _ := cmd.Flags().GetString(FlagNewPayoutAddress)

			msg := &types.MsgEditDataProxy{
				Sender:           clientCtx.GetFromAddress().String(),
				PubKey:           args[0],
				NewMemo:          newMemo,
				NewPayoutAddress: newPayoutAddress,
			}

			feeValue, _ := cmd.Flags().GetString(FlagNewFee)
			if feeValue != "" {
				fee, err := sdk.ParseCoinNormalized(feeValue)
				if err != nil {
					return err
				}

				msg.NewFee = &fee

				feeUpdateDelay, _ := cmd.Flags().GetUint32(FlagFeeUpdateDelay)

				msg.FeeUpdateDelay = feeUpdateDelay
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(FlagMemo, types.DoNotModifyField, "Optionally add a description to the data proxy config")
	cmd.Flags().String(FlagNewPayoutAddress, types.DoNotModifyField, "The new payout address for this data proxy")
	cmd.Flags().String(FlagNewFee, "", "The new fee to be scheduled for this data proxy")
	cmd.Flags().Uint32(FlagFeeUpdateDelay, types.UseMinimumDelay, "Optionally specify a custom delay in blocks. Must be larger than minimum set in module params")

	return cmd
}
