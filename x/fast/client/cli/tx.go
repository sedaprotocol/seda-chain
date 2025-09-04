package cli

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

const (
	FlagMemo         = "memo"
	FlagAdminAddress = "admin-address"
	FlagAddress      = "address"
	FlagNewPubKey    = "new-public-key"
)

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	clientCmd := &cobra.Command{
		Use:                        "client",
		Short:                      "client subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	clientCmd.AddCommand(
		RegisterFastClient(),
		EditFastClient(),
		TransferOwnership(),
		AcceptOwnership(),
		CancelOwnershipTransfer(),
		SettleCredits(),
	)

	userCmd := &cobra.Command{
		Use:                        "user",
		Short:                      "user subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	userCmd.AddCommand(
		AddUser(),
		TopUpUser(),
	)

	cmd.AddCommand(
		clientCmd,
		userCmd,
	)
	return cmd
}

func RegisterFastClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [public_key_hex] [owner_address] [address] --from [authority_address]",
		Short: "Register a fast client.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			ownerAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			adminAddress, err := cmd.Flags().GetString(FlagAdminAddress)
			if err != nil {
				return err
			}
			if adminAddress == "" {
				adminAddress = ownerAddress.String()
			}

			address, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			memo, _ := cmd.Flags().GetString(FlagMemo)

			msg := &types.MsgRegisterFastClient{
				Authority:    clientCtx.GetFromAddress().String(),
				OwnerAddress: ownerAddress.String(),
				AdminAddress: adminAddress,
				Address:      address.String(),
				PublicKey:    pubKeyHex,
				Memo:         memo,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagAdminAddress, "", "The address that can perform administrative actions on the fast client. Defaults to the owner address.")
	cmd.Flags().String(FlagAddress, "", "The address of the fast client that is allowed to submit reports")
	cmd.Flags().String(FlagMemo, "", "Optionally add a description to the fast client")
	return cmd
}

func EditFastClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [public_key_hex] --from [owner_address]",
		Short: "Edit a fast client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			newAdminAddress, _ := cmd.Flags().GetString(FlagAdminAddress)
			newAddress, _ := cmd.Flags().GetString(FlagAddress)
			newPubKey, _ := cmd.Flags().GetString(FlagNewPubKey)
			newMemo, _ := cmd.Flags().GetString(FlagMemo)

			msg := &types.MsgEditFastClient{
				OwnerAddress:        clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
				NewAdminAddress:     newAdminAddress,
				NewAddress:          newAddress,
				NewPublicKey:        newPubKey,
				NewMemo:             newMemo,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(FlagAdminAddress, types.DoNotModifyField, "The new admin address for this fast client")
	cmd.Flags().String(FlagAddress, types.DoNotModifyField, "The new address for this fast client")
	cmd.Flags().String(FlagNewPubKey, types.DoNotModifyField, "The new public key for this fast client")
	cmd.Flags().String(FlagMemo, types.DoNotModifyField, "The new description for the fast client")

	return cmd
}

func TransferOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-ownership [public_key_hex] [new_owner_address] --from [owner_address]",
		Short: "Transfer ownership of a fast client",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			newOwnerAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgTransferOwnership{
				OwnerAddress:        clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
				NewOwnerAddress:     newOwnerAddress.String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func AcceptOwnership() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept-ownership [public_key_hex] --from [new_owner_address]",
		Short: "Accept ownership of a fast client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			msg := &types.MsgAcceptOwnership{
				NewOwnerAddress:     clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CancelOwnershipTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-transfer-ownership [public_key_hex] --from [owner_address]",
		Short: "Cancel ownership transfer of a fast client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			msg := &types.MsgCancelOwnershipTransfer{
				OwnerAddress:        clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func AddUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [fast_client_pubkey_hex] [user_id] [initial_credits] --from [admin_address]",
		Short: "Create a user for a fast client",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			userID := args[1]

			initialCredits := math.NewInt(0)
			if len(args) == 3 {

				parsedCoin, err := sdk.ParseCoinNormalized(args[2])
				if err != nil {
					return err
				}
				initialCredits = parsedCoin.Amount
			}

			msg := &types.MsgAddUser{
				AdminAddress:        clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
				UserId:              userID,
				InitialCredits:      initialCredits,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func TopUpUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-credits [fast_client_pubkey_hex] [user_id] [credits] --from [sender]",
		Short: "Add credits to a fast client's user",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			userID := args[1]

			credits, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgTopUpUser{
				Sender:              clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
				UserId:              userID,
				Amount:              credits.Amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func SettleCredits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settle-credits [public_key_hex] [withdraw|burn] [amount] --from [admin_address]",
		Short: "Settle a fast client's credits",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			settleTypeArg := args[1]
			var settleType types.SettleType
			switch strings.ToLower(settleTypeArg) {
			case "withdraw":
				settleType = types.SETTLE_TYPE_WITHDRAW
			case "burn":
				settleType = types.SETTLE_TYPE_BURN
			default:
				return fmt.Errorf("invalid settle type: %s", settleTypeArg)
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgSettleCredits{
				AdminAddress:        clientCtx.GetFromAddress().String(),
				FastClientPublicKey: pubKeyHex,
				Amount:              amount.Amount,
				SettleType:          settleType,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
