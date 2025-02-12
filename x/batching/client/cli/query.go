package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

const flagWithUnsigned = "with-unsigned"

// GetQueryCmd returns the CLI query commands for batching module.
func GetQueryCmd(_ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryBatch(),
		GetCmdQueryBatchByHeight(),
		GetCmdQueryBatches(),
		GetCmdQueryDataResult(),
	)
	return cmd
}

// GetCmdQueryBatch returns the command for querying a specific batch.
func GetCmdQueryBatch() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch <optional_batch_number>",
		Short: "Query a batch",
		Long:  "Query the latest batch whose signatures have been collected or, by providing its batch number, a specific batch.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var batchNum uint64
			if len(args) != 0 {
				batchNum, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return err
				}
			}
			req := &types.QueryBatchRequest{
				BatchNumber: batchNum,
			}
			res, err := queryClient.Batch(cmd.Context(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryBatchByHeight returns the command for querying a
// batch using a block height.
func GetCmdQueryBatchByHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-by-height <block_height>",
		Short: "Get a batch given its creation block height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			blockHeight, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			req := &types.QueryBatchForHeightRequest{
				BlockHeight: blockHeight,
			}
			res, err := queryClient.BatchForHeight(cmd.Context(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryBatches returns the command for querying all batches.
func GetCmdQueryBatches() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batches",
		Short: "Query all batches",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			withUnsigned, err := cmd.Flags().GetBool(flagWithUnsigned)
			if err != nil {
				return err
			}

			res, err := queryClient.Batches(cmd.Context(),
				&types.QueryBatchesRequest{
					Pagination:   pageReq,
					WithUnsigned: withUnsigned,
				})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool(flagWithUnsigned, false, "include batches without signatures")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "batches")
	return cmd
}

// GetCmdQueryBatch returns the command for querying a data result
// associated with a given data request.
func GetCmdQueryDataResult() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-result <data_request_id> <optional_data_request_height>",
		Short: "Get the data result associated with a given data request",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDataResultRequest{
				DataRequestId: args[0],
			}
			if len(args) == 2 {
				req.DataRequestHeight, err = strconv.ParseUint(args[1], 10, 64)
				if err != nil {
					return err
				}
			}
			res, err := queryClient.DataResult(cmd.Context(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
