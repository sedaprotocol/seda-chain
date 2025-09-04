package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/fast/types"
)

// GetQueryCmd returns the CLI query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetFastClient(),
		GetFastClientTransfer(),
		GetFastClientUsers(),
		GetFastClientUser(),
	)
	return cmd
}

func GetFastClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client [public_key_hex]",
		Short: "Query fast client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			res, err := queryClient.FastClient(cmd.Context(), &types.QueryFastClientRequest{
				FastClientPubKey: pubKeyHex,
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

func GetFastClientTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client-transfer [public_key_hex]",
		Short: "Query pending fast client transfer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			res, err := queryClient.FastClientTransfer(cmd.Context(), &types.QueryFastClientTransferRequest{
				FastClientPubKey: pubKeyHex,
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

func GetFastClientUsers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client-users [public_key_hex]",
		Short: "Query fast client users",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.FastClientUsers(cmd.Context(), &types.QueryFastClientUsersRequest{
				FastClientPubKey: pubKeyHex,
				Pagination:       pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "client-users")
	return cmd
}

func GetFastClientUser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client-user [public_key_hex] [user_id]",
		Short: "Query fast client user",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pubKeyHex := args[0]
			_, err = hex.DecodeString(pubKeyHex)
			if err != nil {
				return err
			}

			userID := args[1]

			res, err := queryClient.FastClientUser(cmd.Context(), &types.QueryFastClientUserRequest{
				FastClientPubKey: pubKeyHex,
				UserId:           userID,
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
		Short: "Query fast module parameters",
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
