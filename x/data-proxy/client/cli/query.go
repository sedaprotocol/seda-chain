package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sedaprotocol/seda-chain/x/data-proxy/types"
)

// GetQueryCmd returns the CLI query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetDataProxyConfig(),
	)
	return cmd
}

// GetDataProxyConfig returns the command for querying the config of a given data proxy.
func GetDataProxyConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-proxy-config <data_proxy_pubkey>",
		Short: "Query a data proxy config by its hex encoded public key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryDataProxyConfigRequest{
				PubKey: args[0],
			}
			res, err := queryClient.DataProxyConfig(cmd.Context(), req)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
