package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// GetQueryCmd returns the CLI query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(GetCmdQueryDataRequestWasm())
	cmd.AddCommand(GetCmdQueryOverlayWasm())
	return cmd
}

// GetCmdQueryDataRequestWasm returns the command for querying Data Request Wasm.
func GetCmdQueryDataRequestWasm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-request-wasm <hash>",
		Short: "Get Data Request Wasm given its hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDataRequestWasmRequest{
				Hash: args[0],
			}
			res, err := queryClient.DataRequestWasm(cmd.Context(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryOverlayWasm returns the command for querying Overlay Wasm.
func GetCmdQueryOverlayWasm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overlay-wasm <hash>",
		Short: "Get Overlay Wasm given its hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOverlayWasmRequest{
				Hash: args[0],
			}
			res, err := queryClient.OverlayWasm(cmd.Context(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
