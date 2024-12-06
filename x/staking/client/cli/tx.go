package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	sdktypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sedaprotocol/seda-chain/app/utils"
	pubkeycli "github.com/sedaprotocol/seda-chain/x/pubkey/client/cli"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	"github.com/sedaprotocol/seda-chain/x/staking/types"
)

const (
	// FlagWithoutSEDAKeys defines a flag to skip generating and
	// uploading SEDA keys.
	FlagWithoutSEDAKeys = "without-seda-keys"
)

// NewTxCmd returns a root CLI command handler for all x/staking transaction commands.
func NewTxCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        sdktypes.ModuleName,
		Short:                      "Staking transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(
		NewCreateValidatorCmd(valAddrCodec),
		stakingcli.NewEditValidatorCmd(valAddrCodec),
		stakingcli.NewDelegateCmd(valAddrCodec, ac),
		stakingcli.NewRedelegateCmd(valAddrCodec, ac),
		stakingcli.NewUnbondCmd(valAddrCodec, ac),
		stakingcli.NewCancelUnbondingDelegation(valAddrCodec, ac),
	)

	return stakingTxCmd
}

func NewCreateValidatorCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-validator [path/to/validator.json]",
		Short: "Create new validator initialized with a self-delegation to it",
		Args:  cobra.ExactArgs(1),
		Long:  `Create a new validator initialized with a self-delegation by submitting a JSON file with the new validator details.`,
		Example: strings.TrimSpace(
			fmt.Sprintf(`
$ %s tx staking create-validator path/to/validator.json --from keyname

Where validator.json contains:

{
	"pubkey": {
		"@type":"/cosmos.crypto.ed25519.PubKey",
		"key":"oWg2ISpLF405Jcm2vXV+2v4fnjodh6aafuIdeoW+rUw="
	},
	"amount": "1000000stake",
	"moniker": "myvalidator",
	"identity": "optional identity signature (ex. UPort or Keybase)",
	"website": "validator's (optional) website",
	"security": "validator's (optional) security contact email",
	"details": "validator's (optional) details",
	"commission-rate": "0.1",
	"commission-max-rate": "0.2",
	"commission-max-change-rate": "0.01",
	"min-self-delegation": "1"
}

where we can get the pubkey using "%s tendermint show-validator"
`, version.AppName, version.AppName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			serverCfg := server.GetServerContextFromCmd(cmd).Config

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			valAddr := sdk.ValAddress(clientCtx.GetFromAddress())
			if valAddr.Empty() {
				return fmt.Errorf("set the from address using --from flag")
			}

			// Generate SEDA keys.
			var pks []pubkeytypes.IndexedPubKey
			withoutSEDAKeys, _ := cmd.Flags().GetBool(FlagWithoutSEDAKeys)
			if !withoutSEDAKeys {
				keyFile, _ := cmd.Flags().GetString(pubkeycli.FlagKeyFile)
				if keyFile != "" {
					pks, err = utils.LoadSEDAPubKeys(keyFile)
					if err != nil {
						return err
					}
				} else {
					pks, err = utils.GenerateSEDAKeys(valAddr, filepath.Dir(serverCfg.PrivValidatorKeyFile()))
					if err != nil {
						return err
					}
				}
			}

			validator, err := parseAndValidateValidatorJSON(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			txf, msg, err := newBuildCreateSEDAValidatorMsg(clientCtx, txf, cmd.Flags(), validator, ac, pks)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().Bool(FlagWithoutSEDAKeys, false, "skip generating and uploading SEDA keys")
	cmd.Flags().String(pubkeycli.FlagKeyFile, "", "path to an existing SEDA key file")
	cmd.Flags().String(stakingcli.FlagIP, "", fmt.Sprintf("The node's public IP. It takes effect only when used in combination with --%s", flags.FlagGenerateOnly))
	cmd.Flags().String(stakingcli.FlagNodeID, "", "The node's ID")
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(flags.FlagFrom)

	return cmd
}

func newBuildCreateSEDAValidatorMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet, val validator, valAc address.Codec, pks []pubkeytypes.IndexedPubKey) (tx.Factory, *types.MsgCreateSEDAValidator, error) {
	valAddr := clientCtx.GetFromAddress()

	description := sdktypes.NewDescription(
		val.Moniker,
		val.Identity,
		val.Website,
		val.Security,
		val.Details,
	)

	valStr, err := valAc.BytesToString(sdk.ValAddress(valAddr))
	if err != nil {
		return txf, nil, err
	}
	msg, err := types.NewMsgCreateSEDAValidator(
		valStr, val.PubKey, pks, val.Amount, description, val.CommissionRates, val.MinSelfDelegation,
	)
	if err != nil {
		return txf, nil, err
	}
	if err := msg.Validate(valAc); err != nil {
		return txf, nil, err
	}

	genOnly, _ := fs.GetBool(flags.FlagGenerateOnly)
	if genOnly {
		ip, _ := fs.GetString(stakingcli.FlagIP)
		p2pPort, _ := fs.GetUint(stakingcli.FlagP2PPort)
		nodeID, _ := fs.GetString(stakingcli.FlagNodeID)

		if nodeID != "" && ip != "" && p2pPort > 0 {
			txf = txf.WithMemo(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort))
		}
	}

	return txf, msg, nil
}

type TxCreateValidatorConfig struct {
	stakingcli.TxCreateValidatorConfig
	SEDAPubKeys []pubkeytypes.IndexedPubKey
}

func BuildCreateSEDAValidatorMsg(clientCtx client.Context, config TxCreateValidatorConfig, txBldr tx.Factory, generateOnly bool, valCodec address.Codec) (tx.Factory, sdk.Msg, error) {
	amounstStr := config.Amount
	amount, err := sdk.ParseCoinNormalized(amounstStr)
	if err != nil {
		return txBldr, nil, err
	}

	valAddr := clientCtx.GetFromAddress()
	description := sdktypes.NewDescription(
		config.Moniker,
		config.Identity,
		config.Website,
		config.SecurityContact,
		config.Details,
	)

	// get the initial validator commission parameters
	rateStr := config.CommissionRate
	maxRateStr := config.CommissionMaxRate
	maxChangeRateStr := config.CommissionMaxChangeRate
	commissionRates, err := buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr)
	if err != nil {
		return txBldr, nil, err
	}

	// get the initial validator min self delegation
	msbStr := config.MinSelfDelegation
	minSelfDelegation, ok := math.NewIntFromString(msbStr)

	if !ok {
		return txBldr, nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "minimum self delegation must be a positive integer")
	}

	valStr, err := valCodec.BytesToString(sdk.ValAddress(valAddr))
	if err != nil {
		return txBldr, nil, err
	}

	msg, err := types.NewMsgCreateSEDAValidator(
		valStr,
		config.PubKey,
		config.SEDAPubKeys,
		amount,
		description,
		commissionRates,
		minSelfDelegation,
	)
	if err != nil {
		return txBldr, msg, err
	}

	if generateOnly {
		ip := config.IP
		p2pPort := config.P2PPort
		nodeID := config.NodeID

		if nodeID != "" && ip != "" && p2pPort > 0 {
			txBldr = txBldr.WithMemo(fmt.Sprintf("%s@%s:%d", nodeID, ip, p2pPort))
		}
	}

	return txBldr, msg, nil
}
