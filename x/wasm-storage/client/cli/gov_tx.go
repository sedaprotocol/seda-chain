package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// DefaultGovAuthority is set to the gov module address.
// Extension point for chains to overwrite the default
var DefaultGovAuthority = sdk.AccAddress(address.Module("gov"))

const (
	flagWasmType  = "wasm-type"
	flagAuthority = "authority"
	flagAmount    = "amount"
	flagLabel     = "label"
	flagAdmin     = "admin"
	flagNoAdmin   = "no-admin"
	flagFixMsg    = "fix-msg"
)

func SubmitProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "submit-proposal",
		Short:        "Submit a wasm proposal.",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		ProposalStoreOverlayCmd(),
		ProposalInstantiateAndRegisterCoreContract(),
	)
	return cmd
}

func ProposalStoreOverlayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-overlay-wasm [wasm file] --wasm-type [wasm_type] --title [text] --summary [text] --authority [address]",
		Short: "Submit a proposal to store a new Overlay Wasm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}
			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			src, err := parseStoreOverlayArgs(args[0], authority, cmd.Flags())
			if err != nil {
				return err
			}

			proposalMsg, err := govv1.NewMsgSubmitProposal([]sdk.Msg{src}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, false)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagWasmType, "", "Overlay Wasm type: data-request-executor or relayer")
	err := cmd.MarkFlagRequired(flagWasmType)
	if err != nil {
		panic(err)
	}
	addCommonProposalFlags(cmd)
	return cmd
}

func ProposalInstantiateAndRegisterCoreContract() *cobra.Command {
	decoder := newArgDecoder(hex.DecodeString)
	cmd := &cobra.Command{
		Use: "instantiate-and-register-core-contract [code_id_int64] [json_encoded_init_args] [salt] --label [text] --admin [address,optional] " +
			"--fix-msg [bool,optional] --title [string] --summary [string] --deposit 10000000aseda",
		Short: "Submit a proposal to instantiate a core contract and register its address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, err := getProposalInfo(cmd)
			if err != nil {
				return err
			}

			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return fmt.Errorf("authority: %s", err)
			}
			if len(authority) == 0 {
				return errors.New("authority address is required")
			}

			src, err := parseInstantiateAndRegisterCoreContractArgs(args[0], args[1], clientCtx.Keyring, authority, cmd.Flags())
			if err != nil {
				return err
			}
			salt, err := decoder.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("salt: %w", err)
			}
			src.Salt = salt

			proposalMsg, err := govv1.NewMsgSubmitProposal([]sdk.Msg{src}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, false)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract during instantiation")
	cmd.Flags().String(flagLabel, "", "A human-readable name for this contract in lists")
	cmd.Flags().String(flagAdmin, "", "Address or key name of an admin")
	cmd.Flags().Bool(flagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	cmd.Flags().Bool(flagFixMsg, false, "An optional flag to include the json_encoded_init_args for the predictable address generation mode")
	decoder.RegisterFlags(cmd.PersistentFlags(), "salt")
	addCommonProposalFlags(cmd)

	return cmd
}

func parseStoreOverlayArgs(file, sender string, _ *flag.FlagSet) (*types.MsgStoreOverlayWasm, error) {
	zipped, err := gzipWasmFile(file)
	if err != nil {
		return nil, err
	}

	msg := &types.MsgStoreOverlayWasm{
		Sender:   sender,
		Wasm:     zipped,
		WasmType: types.WasmTypeFromString(viper.GetString(flagWasmType)),
	}
	return msg, nil
}

func parseInstantiateAndRegisterCoreContractArgs(rawCodeID, initMsg string, kr keyring.Keyring, sender string, flags *flag.FlagSet) (*types.MsgInstantiateAndRegisterCoreContract, error) {
	codeID, err := strconv.ParseUint(rawCodeID, 10, 64)
	if err != nil {
		return nil, err
	}

	amountStr, err := flags.GetString(flagAmount)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	amount, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return nil, fmt.Errorf("amount: %s", err)
	}
	label, err := flags.GetString(flagLabel)
	if err != nil {
		return nil, fmt.Errorf("label: %s", err)
	}
	if label == "" {
		return nil, errors.New("label is required on all contracts")
	}
	adminStr, err := flags.GetString(flagAdmin)
	if err != nil {
		return nil, fmt.Errorf("admin: %s", err)
	}

	noAdmin, err := flags.GetBool(flagNoAdmin)
	if err != nil {
		return nil, fmt.Errorf("no-admin: %s", err)
	}

	// ensure sensible admin is set (or explicitly immutable)
	if adminStr == "" && !noAdmin {
		return nil, fmt.Errorf("you must set an admin or explicitly pass --no-admin to make it immutible (wasmd issue #719)")
	}
	if adminStr != "" && noAdmin {
		return nil, fmt.Errorf("you set an admin and passed --no-admin, those cannot both be true")
	}

	if adminStr != "" {
		addr, err := sdk.AccAddressFromBech32(adminStr)
		if err != nil {
			info, err := kr.Key(adminStr)
			if err != nil {
				return nil, fmt.Errorf("admin %s", err)
			}
			admin, err := info.GetAddress()
			if err != nil {
				return nil, err
			}
			adminStr = admin.String()
		} else {
			adminStr = addr.String()
		}
	}

	fixMsg, err := flags.GetBool(flagFixMsg)
	if err != nil {
		return nil, fmt.Errorf("fix msg: %w", err)
	}

	msg := types.MsgInstantiateAndRegisterCoreContract{
		Sender: sender,
		CodeID: codeID,
		Label:  label,
		Funds:  amount,
		Msg:    []byte(initMsg),
		Admin:  adminStr,
		FixMsg: fixMsg,
	}
	return &msg, nil
}

func addCommonProposalFlags(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(cli.FlagTitle, "", "Title of proposal")
	cmd.Flags().String(cli.FlagSummary, "", "Summary of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "Deposit of proposal")
	cmd.Flags().String(flagAuthority, DefaultGovAuthority.String(), "The address of the governance account. Default is the sdk gov module account")
}

func getProposalInfo(cmd *cobra.Command) (client.Context, string, string, sdk.Coins, error) {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return client.Context{}, "", "", nil, err
	}

	proposalTitle, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return clientCtx, proposalTitle, "", nil, err
	}

	summary, err := cmd.Flags().GetString(cli.FlagSummary)
	if err != nil {
		return client.Context{}, proposalTitle, summary, nil, err
	}

	depositArg, err := cmd.Flags().GetString(cli.FlagDeposit)
	if err != nil {
		return client.Context{}, proposalTitle, summary, nil, err
	}

	deposit, err := sdk.ParseCoinsNormalized(depositArg)
	if err != nil {
		return client.Context{}, proposalTitle, summary, deposit, err
	}

	return clientCtx, proposalTitle, summary, deposit, nil
}
