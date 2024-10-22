package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
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

	cmd.AddCommand(
		GetCmdStoreOracleProgram(),
		SubmitProposalCmd(),
	)
	return cmd
}

// GetCmdStoreOracleProgram returns the command for storing a
// oracle program file.
func GetCmdStoreOracleProgram() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-oracle-program [wasm file]",
		Short: "Store oracle program file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			wasm, err := gzipWasmFile(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgStoreOracleProgram{
				Sender: clientCtx.GetFromAddress().String(),
				Wasm:   wasm,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func gzipWasmFile(filename string) ([]byte, error) {
	wasm, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	if !ioutils.IsWasm(wasm) {
		return nil, fmt.Errorf("invalid Wasm file")
	}

	zipped, err := ioutils.GzipIt(wasm)
	if err != nil {
		return nil, err
	}
	return zipped, nil
}
