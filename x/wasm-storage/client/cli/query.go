package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// GetQueryCmd returns the CLI query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryDataRequestWasm(),
		GetCmdQueryExecutorWasm(),
		GetCmdQueryDataRequestWasms(),
		GetCmdQueryExecutorWasms(),
		GetCmdQueryCoreContractRegistry(),
	)
	return cmd
}

// GetCmdQueryDataRequestWasm returns the command for querying data request wasm..
func GetCmdQueryDataRequestWasm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-request-wasm <hash>",
		Short: "Get data request wasm given its hash",
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

// GetCmdQueryExecutorWasm returns the command for querying an executor Wasm.
func GetCmdQueryExecutorWasm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executor-wasm <hash>",
		Short: "Get an executor wasm given its hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryExecutorWasmRequest{
				Hash: args[0],
			}
			res, err := queryClient.ExecutorWasm(cmd.Context(), req)
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
// data request wasms in the store.
func GetCmdQueryDataRequestWasms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-data-request-wasms",
		Short: "List hashes and expiration heights of all data request wasms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

// GetCmdQueryExecutorWasms returns the command for querying all
// executor wasms.
func GetCmdQueryExecutorWasms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-executor-wasms",
		Short: "List hashes of all executor wasms",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ExecutorWasms(cmd.Context(), &types.QueryExecutorWasmsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCoreContractRegistry returns the command for querying
// Core Contract registry.
func GetCmdQueryCoreContractRegistry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "core-contract-registry",
		Short: "Get the address of Core Contract",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CoreContractRegistry(
				cmd.Context(),
				&types.QueryCoreContractRegistryRequest{},
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
