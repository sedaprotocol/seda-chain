package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/core/types"
)

// GetQueryCmd returns the CLI query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetOwner(),
		GetPendingOwner(),
		GetPaused(),
		GetAllowlist(),
		GetStaker(),
		GetExecutors(),
		GetDataRequest(),
		GetCmdQueryParams(),
		GetStakingConfig(),
		GetDataRequestConfig(),
	)
	return cmd
}

func GetOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "owner",
		Short: "Query the owner of the core module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Owner(cmd.Context(), &types.QueryOwnerRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetPendingOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-owner",
		Short: "Query the pending owner of the core module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PendingOwner(cmd.Context(), &types.QueryPendingOwnerRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetPaused() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paused",
		Short: "Query whether the core module is paused",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Paused(cmd.Context(), &types.QueryPausedRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetAllowlist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allowlist",
		Short: "Query the executor allowlist",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Allowlist(cmd.Context(), &types.QueryAllowlistRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetStaker() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "staker [public_key]",
		Short: "Query the staker info for a given executor public key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Staker(cmd.Context(), &types.QueryStakerRequest{
				PublicKey: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetExecutors() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executors",
		Short: "Query the list of executors",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			offset, err := cmd.Flags().GetUint32(flags.FlagOffset)
			if err != nil {
				return err
			}
			limit, err := cmd.Flags().GetUint32(flags.FlagLimit)
			if err != nil {
				return err
			}

			res, err := queryClient.Executors(cmd.Context(), &types.QueryExecutorsRequest{
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint32(flags.FlagOffset, 0, fmt.Sprintf("pagination offset of executors to query for"))
	cmd.Flags().Uint32(flags.FlagLimit, 100, fmt.Sprintf("pagination limit of executors to query for"))
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetDataRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-request [dr_id]",
		Short: "Query a data request by its ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DataRequest(cmd.Context(), &types.QueryDataRequestRequest{
				DrId: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query core module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetStakingConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "staking-config",
		Short: "Query the staking configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.StakingConfig(cmd.Context(), &types.QueryStakingConfigRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetDataRequestConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-request-config",
		Short: "Query the data request configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DataRequestConfig(cmd.Context(), &types.QueryDataRequestConfigRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
