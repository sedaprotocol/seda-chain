package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// DefaultGovAuthority is set to the gov module address.
// Extension point for chains to overwrite the default
var DefaultGovAuthority = sdk.AccAddress(address.Module("gov"))

const flagAuthority = "authority"

func SubmitProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "submit-proposal",
		Short:        "Submit a wasm proposal.",
		SilenceUsage: true,
	}
	cmd.AddCommand(
		ProposalStoreCodeCmd(),
	)
	return cmd
}

func ProposalStoreCodeCmd() *cobra.Command {
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

			src, err := parseStoreCodeArgs(args[0], authority, cmd.Flags())
			if err != nil {
				return err
			}

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{&src}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary)
			if err != nil {
				return err
			}
			if err = proposalMsg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(FlagWasmType, "", "Overlay Wasm type: data-request-executor or relayer")
	cmd.MarkFlagRequired(FlagWasmType)
	// proposal flags
	addCommonProposalFlags(cmd)
	return cmd
}

func parseStoreCodeArgs(file string, sender string, flags *flag.FlagSet) (types.MsgStoreOverlayWasm, error) {
	zipped, err := gzipWasmFile(file)
	if err != nil {
		return types.MsgStoreOverlayWasm{}, err
	}

	msg := types.MsgStoreOverlayWasm{
		Sender:   sender,
		Wasm:     zipped,
		WasmType: types.WasmTypeFromString(viper.GetString(FlagWasmType)),
	}
	return msg, nil
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
