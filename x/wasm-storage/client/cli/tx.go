package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
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
		Use:   "store-oracle-program [wasm_file]",
		Short: "Store an oracle program",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get the params to check the max WASM size and cost per byte.
			queryClient := types.NewQueryClient(clientCtx)
			params, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			length, wasm, err := gzipWasmFile(args[0])
			if err != nil {
				return err
			}

			if length > params.Params.MaxWasmSize {
				return fmt.Errorf("WASM file is too large. Max size is %d bytes", params.Params.MaxWasmSize)
			}

			storageFee := math.NewIntFromUint64(params.Params.WasmCostPerByte).Mul(math.NewInt(length))

			msg := &types.MsgStoreOracleProgram{
				Sender:     clientCtx.GetFromAddress().String(),
				Wasm:       wasm,
				StorageFee: sdk.NewCoins(sdk.NewCoin(appparams.DefaultBondDenom, storageFee)),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// gzipWasmFile returns the length of the unzipped wasm file and the zipped wasm file.
func gzipWasmFile(filename string) (int64, []byte, error) {
	wasm, err := os.ReadFile(filename)
	if err != nil {
		return 0, nil, err
	}

	if !ioutils.IsWasm(wasm) {
		return 0, nil, fmt.Errorf("invalid Wasm file")
	}

	zipped, err := ioutils.GzipIt(wasm)
	if err != nil {
		return 0, nil, err
	}
	return int64(len(wasm)), zipped, nil
}
