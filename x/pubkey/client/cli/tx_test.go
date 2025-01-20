package cli_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/pubkey/client/cli"
)

type CLITestSuite struct {
	suite.Suite

	kr          keyring.Keyring
	encCfg      testutilmod.TestEncodingConfig
	baseCtx     client.Context
	clientCtx   client.Context
	commonFlags []string
	saveDir     string
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(params.Bech32PrefixAccAddr, params.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(params.Bech32PrefixValAddr, params.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(params.Bech32PrefixConsAddr, params.Bech32PrefixConsPub)
	config.Seal()

	s.encCfg = testutilmod.MakeTestEncodingConfig()
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen()

	// Create a new account and set it as from flag
	info, _, err := s.clientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pk, err := info.GetPubKey()
	s.Require().NoError(err)
	account := sdk.AccAddress(pk.Address())

	s.commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("seda", sdkmath.NewInt(10))).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, account.String()),
		// Generating a new encryption key or manually setting the encryption key
		// requires user interaction, so we disable it here for testing
		fmt.Sprintf("--%s", cli.FlagNoEncryption),
	}
}

func (s *CLITestSuite) TestAddSEDAKeys() {
	cmd := cli.AddKey(s.clientCtx.Codec.InterfaceRegistry().SigningContext().ValidatorAddressCodec())
	cmd.SetOutput(io.Discard)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"generate new keys",
			s.commonFlags,
			"",
		},
		{
			"key-file flag, nonexistent file",
			append([]string{fmt.Sprintf("--%s=%s", cli.FlagKeyFile, "./config/seda_keyys.json")}, s.commonFlags...),
			"failed to read SEDA keys from ./config/seda_keyys.json: open ./config/seda_keyys.json: no such file or directory",
		},
		{
			"key-file flag, valid file",
			append([]string{fmt.Sprintf("--%s=%s", cli.FlagKeyFile, "./config/seda_keys.json")}, s.commonFlags...),
			"",
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(s.baseCtx, cmd))
			s.saveDir = filepath.Dir(server.GetServerContextFromCmd(cmd).Config.PrivValidatorKeyFile())

			out, err := clitestutil.ExecTestCLICmd(s.baseCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				msg := &sdk.TxResponse{}
				s.Require().NoError(s.baseCtx.Codec.UnmarshalJSON(out.Bytes(), msg), out.String())
			}
		})
	}
}

func (s *CLITestSuite) TearDownSuite() {
	err := os.RemoveAll(s.saveDir)
	s.Require().NoError(err)
}
