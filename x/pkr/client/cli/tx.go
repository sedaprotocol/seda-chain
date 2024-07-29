package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/sedaprotocol/seda-chain/app/utils"
	"github.com/sedaprotocol/seda-chain/x/pkr/types"
)

// GetTxCmd returns the CLI transaction commands for this module
func GetTxCmd(valAddrCodec address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(
		AddKey(valAddrCodec),
	)
	return cmd
}

// AddKey returns the command for adding a new key and uploading its
// public key on chain at a given index.
func AddKey(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-key [index]",
		Short: "Create a new key and upload its public key on chain at a given index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			serverCtx := server.GetServerContextFromCmd(cmd)

			valAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			index, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return errorsmod.Wrap(fmt.Errorf("invalid index: %d", index), "invalid index")
			}
			var pk crypto.PubKey
			switch index {
			case 0:
				// VRF key derived using secp256k1
				pk, err = utils.InitializeVRFKey(serverCtx.Config, "vrf_key") // TODO: remove key name arg
				if err != nil {
					return errorsmod.Wrap(err, "failed to initialize a new key")
				}
			default:
				panic("unsupported index")
			}

			pkAny, err := codectypes.NewAnyWithValue(pk)
			if err != nil {
				return err
			}
			msg := &types.MsgAddKey{
				ValidatorAddress: valAddr,
				Index:            uint32(index),
				Pubkey:           pkAny,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
