package cli

import (
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const FlagWasmType = "wasm-type"

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(GetCmdStoreDataRequestWasm())
	cmd.AddCommand(GetCmdStoreOverlayWasm())
	return cmd
}

// GetCmdStoreDataRequestWasm returns the command for storing Data Request Wasm.
func GetCmdStoreDataRequestWasm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-data-request-wasm <filename>",
		Short: "Store Data Request Wasm file",
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

			msg := &types.MsgStoreDataRequestWasm{
				Sender:   clientCtx.GetFromAddress().String(),
				Wasm:     wasm,
				WasmType: types.WasmTypeFromString(viper.GetString(FlagWasmType)),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagWasmType, "", "Wasm type")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdStoreOverlayWasm returns the command for storing Overlay Wasm.
func GetCmdStoreOverlayWasm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-overlay-wasm <filename>",
		Short: "Store Overlay Wasm file",
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

			msg := &types.MsgStoreOverlayWasm{
				Sender:   clientCtx.GetFromAddress().String(),
				Wasm:     wasm,
				WasmType: types.WasmTypeFromString(viper.GetString(FlagWasmType)),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagWasmType, "", "Wasm type")
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
