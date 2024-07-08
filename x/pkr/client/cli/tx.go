package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/sedaprotocol/seda-chain/app/utils"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		AddVRFKey(),
	)
	return cmd
}

// AddVRFKey adds a VRF key.
func AddVRFKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name] [application]",
		Short: "Add an application specific encrypted private key",
		Long: `Derive a new private key and encrypt to disk.
    pkr add vrf_for_randomness
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)
			pk, err := utils.InitializeVRFKey(serverCtx.Config, args[0])
			if err != nil {
				return errorsmod.Wrap(err, "failed to initialize VRF key")
			}

			pkAny, err := codectypes.NewAnyWithValue(pk)
			if err != nil {
				return err
			}
			msg := &types.MsgAddVRFKey{
				Name:        args[0],
				Application: args[1],
				Pubkey:      pkAny,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	return cmd
}
