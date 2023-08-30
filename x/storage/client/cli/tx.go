package cli

import (
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sedaprotocol/seda-chain/x/storage/types"
)

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(GetCmdStore())
	return cmd
}

// GetCmdStore returns the store tx.
func GetCmdStore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store <filename>",
		Short: "Store Wasm or gzip file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			wasm, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			// gzip the wasm file
			if ioutils.IsWasm(wasm) {
				wasm, err = ioutils.GzipIt(wasm)
				if err != nil {
					return err
				}
			} else if !ioutils.IsGzip(wasm) {
				return fmt.Errorf("invalid input file. Use wasm binary or gzip")
			}

			msg := &types.MsgStore{
				Sender: clientCtx.GetFromAddress().String(),
				Data:   wasm,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
