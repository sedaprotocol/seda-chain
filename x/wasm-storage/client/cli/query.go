package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// GetQueryCmd returns the CLI query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryDataRequestWasm(),
		GetCmdQueryOverlayWasm(),
		GetCmdQueryDataRequestWasms(),
		GetCmdQueryOverlayWasms(),
		GetCmdQueryProxyContractRegistry(),
	)
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

// GetCmdQueryDataRequestWasms returns the command for querying
// hashes and types of all Data Request Wasms.
func GetCmdQueryDataRequestWasms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-data-request-wasms",
		Short: "List hashes and types of all Data Request Wasms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DataRequestWasms(cmd.Context(), &types.QueryDataRequestWasmsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryOverlayWasms returns the command for querying
// hashes and types of all Overlay Wasms.
func GetCmdQueryOverlayWasms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-overlay-wasms",
		Short: "List hashes and types of all Overlay Wasms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.OverlayWasms(cmd.Context(), &types.QueryOverlayWasmsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryProxyContractRegistry returns the command for querying
// Proxy Contract registry.
func GetCmdQueryProxyContractRegistry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy-contract-registry",
		Short: "Get the address of Proxy Contract",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ProxyContractRegistry(
				cmd.Context(),
				&types.QueryProxyContractRegistryRequest{},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
