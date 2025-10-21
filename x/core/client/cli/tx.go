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

const (
	FlagWithdrawAddress   = "withdraw-address"
	FlagMemo              = "memo"
	FlagVersion           = "version"
	FlagReplicationFactor = "replication-factor"
	FlagGasPrice          = "gas-price"
	FlagExecProgramID     = "exec-program-id"
	FlagExecGasLimit      = "exec-gas-limit"
	FlagTallyProgramID    = "tally-program-id"
	FlagTallyGasLimit     = "tally-gas-limit"
	FlagConsensusFilter   = "consensus-filter"
	FlagExecInputs        = "exec-inputs"
	FlagTallyInputs       = "tally-inputs"
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
		Stake(),
		Unstake(),
		Withdraw(),
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
		Short: "Add an executor public key to the allowlist",
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
		Short: "Remove an executor public key from the allowlist",
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

func Stake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake [public_key_hex] [proof_hex] [stake] --memo [memo_base64] --from [staker]",
		Short: "Stake for the given staker public key",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Public key and proof must be valid hex strings.
			publicKey := args[0]
			proof := args[1]

			_, err = hex.DecodeString(publicKey)
			if err != nil {
				return err
			}
			_, err = hex.DecodeString(proof)
			if err != nil {
				return err
			}

			// Memo, if provided, must be a valid base64 string.
			memo, err := cmd.Flags().GetString("memo")
			if err != nil {
				return err
			}
			if memo != "" {
				_, err = base64.StdEncoding.DecodeString(memo)
				if err != nil {
					return err
				}
			}

			// Stake must be a valid coin.
			stake, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgStake{
				Sender:    clientCtx.GetFromAddress().String(),
				PublicKey: publicKey,
				Memo:      memo,
				Proof:     proof,
				Stake:     stake,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringP("memo", "m", "", "optional memo for staking")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func Unstake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake [public_key_hex] [proof_hex] --from [staker]",
		Short: "Unstake all staked tokens for the given staker public key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Public key and proof must be valid hex strings.
			publicKey := args[0]
			proof := args[1]

			_, err = hex.DecodeString(publicKey)
			if err != nil {
				return err
			}
			_, err = hex.DecodeString(proof)
			if err != nil {
				return err
			}

			msg := &types.MsgUnstake{
				Sender:    clientCtx.GetFromAddress().String(),
				PublicKey: publicKey,
				Proof:     proof,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func Withdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [public_key_hex] [proof_hex] --from [staker]",
		Short: "Withdraw rewards for the given staker public key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Public key and proof must be valid hex strings.
			publicKey := args[0]
			proof := args[1]

			_, err = hex.DecodeString(publicKey)
			if err != nil {
				return err
			}
			_, err = hex.DecodeString(proof)
			if err != nil {
				return err
			}

			withdrawAddress, err := cmd.Flags().GetString(FlagWithdrawAddress)
			if err != nil {
				return err
			}
			if withdrawAddress == "" {
				withdrawAddress = clientCtx.GetFromAddress().String()
			}

			msg := &types.MsgWithdraw{
				Sender:          clientCtx.GetFromAddress().String(),
				PublicKey:       publicKey,
				Proof:           proof,
				WithdrawAddress: withdrawAddress,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagWithdrawAddress, "", "optional withdraw address (defaults to tx sender)")
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

			if rf, err := cmd.Flags().GetUint32(FlagReplicationFactor); err == nil {
				if rf == 0 {
					return fmt.Errorf("replication factor must be greater than 0")
				}
				msg.ReplicationFactor = rf
			}

			if version, err := cmd.Flags().GetString(FlagVersion); err == nil {
				if version != "" {
					msg.Version = version
				}
			}

			if memo, err := cmd.Flags().GetString(FlagMemo); err == nil {
				if memo != "" {
					msg.Memo = []byte(base64.StdEncoding.EncodeToString([]byte(memo)))
				}
			}

			execID, err := cmd.Flags().GetString(FlagExecProgramID)
			if err != nil {
				return err
			}
			if execID == "" {
				return fmt.Errorf("exec-program-id is required")
			}
			msg.ExecProgramID = execID

			execInputs, err := cmd.Flags().GetString(FlagExecInputs)
			if err != nil {
				return err
			}
			if _, err = hex.DecodeString(execInputs); err != nil {
				return fmt.Errorf("exec-inputs must be a valid hex string")
			}
			msg.ExecInputs = []byte(execInputs)

			tallyID, err := cmd.Flags().GetString(FlagTallyProgramID)
			if err != nil {
				return err
			}
			if tallyID == "" {
				msg.TallyProgramID = msg.ExecProgramID
			} else {
				msg.TallyProgramID = tallyID
			}

			tallyInputs, err := cmd.Flags().GetString(FlagTallyInputs)
			if err != nil {
				return err
			}
			if _, err = hex.DecodeString(tallyInputs); err != nil {
				return fmt.Errorf("tally-inputs must be a valid hex string")
			}
			msg.TallyInputs = []byte(tallyInputs)

			gasPriceString, err := cmd.Flags().GetString(FlagGasPrice)
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

			execGasLimit, err := cmd.Flags().GetUint64(FlagExecGasLimit)
			if err != nil {
				return err
			}
			msg.ExecGasLimit = execGasLimit

			tallyGasLimit, err := cmd.Flags().GetUint64(FlagTallyGasLimit)
			if err != nil {
				return err
			}
			msg.TallyGasLimit = tallyGasLimit

			consensusFilter, err := cmd.Flags().GetString(FlagConsensusFilter)
			if err != nil {
				return err
			}
			bytes, err := hex.DecodeString(consensusFilter)
			if err != nil {
				return fmt.Errorf("consensus-filter must be a valid hex string")
			}
			msg.ConsensusFilter = bytes

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringP(FlagMemo, "m", "", "optional memo for the data request")
	cmd.Flags().StringP(FlagVersion, "v", "v1.0.0", "version for the data request")
	cmd.Flags().Uint32P(FlagReplicationFactor, "r", 1, "replication factor for the data request")
	cmd.Flags().String(FlagGasPrice, "2000", "gas price for the data request")
	cmd.Flags().StringP(FlagExecProgramID, "e", "", "execution program ID for the data request in hex")
	cmd.Flags().Uint64(FlagExecGasLimit, 300_000_000_000_000, "execution gas limit for the data request")
	cmd.Flags().StringP(FlagTallyProgramID, "t", "", "tally program ID for the data request")
	cmd.Flags().Uint64(FlagTallyGasLimit, 300_000_000_000_000, "tally gas limit for the data request")
	cmd.Flags().String(FlagExecInputs, "", "execution inputs for the data request encoded in hex")
	cmd.Flags().String(FlagTallyInputs, "", "tally inputs for the data request encoded in hex")
	cmd.Flags().StringP(FlagConsensusFilter, "", "", "optional consensus filter for the data request encoded in hex")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
