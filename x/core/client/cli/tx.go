package cli

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	cosmossdk_io_math "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		PostDataRequest(),
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
		Use:   "post-data-request [public_key] [funds] --from [requester_address]",
		Short: "Post a data request to the core module",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if _, err := hex.DecodeString(args[0]); err != nil {
				return fmt.Errorf("public key must be a valid hex string")
			}

			funds, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}
			if len(funds) != 1 {
				return fmt.Errorf("must provide exactly one denomination of funds")
			}
			if !funds[0].IsValid() || funds[0].IsZero() {
				return fmt.Errorf("funds must be a valid, non-zero amount")
			}
			if funds[0].Denom != "aseda" {
				return fmt.Errorf("funds must be in aseda")
			}

			msg := &types.MsgPostDataRequest{
				Sender: clientCtx.GetFromAddress().String(),
				Funds:  funds[0],
			}

			if rf, err := cmd.Flags().GetUint32("replication-factor"); err == nil {
				if rf == 0 {
					return fmt.Errorf("replication factor must be greater than 0")
				}
				msg.ReplicationFactor = rf
			}

			if version, err := cmd.Flags().GetString("version"); err == nil {
				if version != "" {
					msg.Version = version
				}
			}

			if memo, err := cmd.Flags().GetString("memo"); err == nil {
				if memo != "" {
					msg.Memo = []byte(base64.StdEncoding.EncodeToString([]byte(memo)))
				}
			}

			execID, err := cmd.Flags().GetString("exec-program-id")
			if err != nil {
				return err
			}
			if execID == "" {
				return fmt.Errorf("exec-program-id is required")
			}
			msg.ExecProgramID = execID

			execInputs, err := cmd.Flags().GetString("exec-inputs")
			if err != nil {
				return err
			}
			msg.ExecInputs = []byte(execInputs)

			tallyID, err := cmd.Flags().GetString("tally-program-id")
			if err != nil {
				return err
			}
			if tallyID == "" {
				msg.TallyProgramID = msg.ExecProgramID
			} else {
				msg.TallyProgramID = tallyID
			}

			tallyInputs, err := cmd.Flags().GetString("tally-inputs")
			if err != nil {
				return err
			}
			msg.TallyInputs = []byte(tallyInputs)

			gasPriceString, err := cmd.Flags().GetString("gas-price")
			if err != nil {
				return err
			}
			gasPrice, ok := cosmossdk_io_math.NewIntFromString(gasPriceString)
			if !ok {
				return fmt.Errorf("gas price must be a valid integer")
			}
			if gasPrice.LT(cosmossdk_io_math.NewInt(2000)) {
				return fmt.Errorf("gas price must be at least 2000")
			}
			msg.GasPrice = gasPrice

			execGasLimit, err := cmd.Flags().GetUint64("exec-gas-limit")
			if err != nil {
				return err
			}
			msg.ExecGasLimit = execGasLimit

			tallyGasLimit, err := cmd.Flags().GetUint64("tally-gas-limit")
			if err != nil {
				return err
			}
			msg.TallyGasLimit = tallyGasLimit

			consensusFilter, err := cmd.Flags().GetString("consensus-filter")
			if err != nil {
				return err
			}

			switch consensusFilter {
			case "":
				break
			case "mad":
				msg.ConsensusFilter = []byte("mad")
			case "mode":
				msg.ConsensusFilter = []byte("mode")
			default:
				return fmt.Errorf("invalid consensus filter: %s", consensusFilter)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringP("memo", "m", "", "optional memo for the data request")
	cmd.Flags().StringP("version", "v", "v1.0.0", "version for the data request")
	cmd.Flags().Uint32P("replication-factor", "r", 1, "replication factor for the data request")
	cmd.Flags().String("gas-price", "2000", "gas price for the data request")
	cmd.Flags().StringP("exec-program-id", "e", "", "execution program ID for the data request")
	cmd.Flags().Uint64("exec-gas-limit", 300_000_000_000_000, "execution gas limit for the data request")
	cmd.Flags().StringP("tally-program-id", "t", "", "tally program ID for the data request")
	cmd.Flags().Uint64("tally-gas-limit", 300_000_000_000_000, "tally gas limit for the data request")
	cmd.Flags().String("exec-inputs", "", "execution inputs for the data request")
	cmd.Flags().String("tally-inputs", "", "tally inputs for the data request")
	cmd.Flags().StringP("consensus-filter", "", "", "optional consensus filter for the data request (options: mad, mode)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
