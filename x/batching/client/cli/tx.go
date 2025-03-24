package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	evidencetypes "cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sedatypes "github.com/sedaprotocol/seda-chain/types"
	"github.com/sedaprotocol/seda-chain/x/batching/types"
)

const (
	// FlagProvingScheme defines a flag to specify the proving scheme.
	FlagProvingScheme = "proving-scheme"
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
		SubmitBatchDoubleSign(),
	)
	return cmd
}

func SubmitBatchDoubleSign() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-double-sign [batch_height] [block_height] [operator_address] [validator_root] [data_result_root] [proving_metadata_hash] [signature]",
		Short: "Submit evidence of a validator double signing a batch",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			submitter := clientCtx.GetFromAddress().String()
			if submitter == "" {
				return fmt.Errorf("set the from address using --from flag")
			}

			batchNumber, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid batch number: %s", args[0])
			}

			blockHeight, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid block height: %s", args[1])
			}

			operatorAddr := args[2]
			valAddr := sdk.ValAddress(operatorAddr)
			if valAddr.Empty() {
				return fmt.Errorf("invalid operator address: %s", args[2])
			}

			validatorRoot := args[3]
			if validatorRoot == "" {
				return fmt.Errorf("invalid validator root: %s", args[3])
			}

			dataResultRoot := args[4]
			if dataResultRoot == "" {
				return fmt.Errorf("invalid data result root: %s", args[4])
			}

			provingMetadataHash := args[5]
			if provingMetadataHash == "" {
				return fmt.Errorf("invalid proving metadata hash: %s", args[5])
			}

			signature := args[6]
			if signature == "" {
				return fmt.Errorf("invalid signature: %s", args[6])
			}

			// It's easier to use a uint64 as it's the return type of the strconv.ParseUint function
			var provingSchemeIndex uint64
			provingSchemeInput, _ := cmd.Flags().GetString(FlagProvingScheme)
			if provingSchemeInput != "" {
				provingSchemeIndex, err = strconv.ParseUint(provingSchemeInput, 10, 32)
				if err != nil || provingSchemeIndex != uint64(sedatypes.SEDAKeyIndexSecp256k1) {
					return fmt.Errorf("invalid proving scheme index: %s", provingSchemeInput)
				}
			} else {
				provingSchemeIndex = uint64(sedatypes.SEDAKeyIndexSecp256k1)
			}

			evidence := &types.BatchDoubleSign{
				BatchNumber:         batchNumber,
				BlockHeight:         blockHeight,
				OperatorAddress:     operatorAddr,
				ValidatorRoot:       validatorRoot,
				DataResultRoot:      dataResultRoot,
				ProvingMetadataHash: provingMetadataHash,
				Signature:           signature,
				ProvingSchemeIndex:  uint32(provingSchemeIndex),
			}

			evidencePacked, err := codectypes.NewAnyWithValue(evidence)
			if err != nil {
				return err
			}

			msg := &evidencetypes.MsgSubmitEvidence{
				Submitter: submitter,
				Evidence:  evidencePacked,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagProvingScheme, "0", fmt.Sprintf("proving scheme index [%d]", sedatypes.SEDAKeyIndexSecp256k1))
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
