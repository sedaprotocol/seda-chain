package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	ibclocalhost "github.com/cosmos/ibc-go/v8/modules/light-clients/09-localhost"

	"cosmossdk.io/math"

	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/sedaprotocol/seda-chain/interchaintest/types"
)

var (
	/* =================================================== */
	/*                   CHAIN CONFIG                    */
	/* =================================================== */
	coinType      = "118"
	SedaDenom     = "aseda"
	SedaChainName = "seda"

	dockerImage = ibc.DockerImage{
		Repository: "sedad-e2e", // FOR LOCAL IMAGE USE: Docker Image Name
		Version:    "latest",    // FOR LOCAL IMAGE USE: Docker Image Tag
		UidGid:     "1025:1025",
	}

	SedaRepo = "ghcr.io/sedaprotocol/seda-chain"

	coinDecimals int64 = 18

	SedaCfg = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                SedaChainName,
		ChainID:             "seda-local-1",
		Images:              []ibc.DockerImage{dockerImage},
		Bin:                 "sedad",
		Bech32Prefix:        "seda",
		Denom:               SedaDenom,
		CoinType:            coinType,
		CoinDecimals:        &coinDecimals,
		GasPrices:           fmt.Sprintf("0%s", SedaDenom),
		GasAdjustment:       2.0,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		SkipGenTx:           false,
		PreGenesis:          nil,
		EncodingConfig:      sedaEncoding(),
		ModifyGenesis:       nil,
		ConfigFileOverrides: nil,
	}

	/* =================================================== */
	/*                   RELAYER CONFIG                    */
	/* =================================================== */
	RlyConfig = types.RelayerConfig{
		Type:    ibc.CosmosRly,
		Name:    "relay",
		Image:   "ghcr.io/cosmos/relayer",
		Version: "main",
	}

	/* =================================================== */
	/*                     GOV CONFIG                      */
	/* =================================================== */
	VotingPeriod               = "15s"
	MaxDepositPeriod           = "10s"
	VoteExtensionsEnableHeight = "100000" // essentially disabled

	/* =================================================== */
	/*                    WALLET CONFIG                    */
	/* =================================================== */
	GenesisWalletAmount = math.NewInt(10_000_000_000)
)

// sedaEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func sedaEncoding() *sdktestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	ibclocalhost.RegisterInterfaces(cfg.InterfaceRegistry)
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func GetTestGenesis() []cosmos.GenesisKV {
	return []cosmos.GenesisKV{
		{
			Key:   "app_state.gov.params.voting_period",
			Value: VotingPeriod,
		},
		{
			Key:   "app_state.gov.params.max_deposit_period",
			Value: MaxDepositPeriod,
		},
		{
			Key:   "app_state.gov.params.min_deposit.0.denom",
			Value: SedaDenom,
		},
		{
			Key:   "consensus.params.abci.vote_extensions_enable_height",
			Value: VoteExtensionsEnableHeight,
		},
	}
}

func GetSEDAConfig() ibc.ChainConfig {
	cfg := SedaCfg
	cfg.ModifyGenesis = cosmos.ModifyGenesis(GetTestGenesis())
	cfg.ConfigFileOverrides = GetSEDAAppTomlOverrides()
	return cfg
}

func GetSEDAAppTomlOverrides() map[string]any {
	appTomlOverrides := make(testutil.Toml)

	appTomlOverrides["seda.enable-seda-signer"] = true
	appTomlOverrides["seda.seda-key-file"] = "./config/seda_keys.json"
	appTomlOverrides["seda.allow-unencrypted-seda-keys"] = true

	configFileOverrides := make(map[string]any)
	configFileOverrides["config/app.toml"] = appTomlOverrides
	return configFileOverrides
}

// CreateChains generates this branch's chain (ex: from the commit)
func CreateChains(t *testing.T, numVals, numFullNodes int, configFileOverrides map[string]any) []ibc.Chain {
	t.Helper()
	cfg := SedaCfg
	cfg.ModifyGenesis = cosmos.ModifyGenesis(GetTestGenesis())
	cfg.Images = []ibc.DockerImage{dockerImage}
	cfg.ConfigFileOverrides = configFileOverrides
	return CreateChainsWithCustomConfig(t, numVals, numFullNodes, cfg)
}

func CreateChainsWithCustomConfig(t *testing.T, numVals, numFullNodes int, config ibc.ChainConfig) []ibc.Chain {
	t.Helper()

	logger := zaptest.NewLogger(t)
	cf := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		{
			Name:          SedaCfg.Name,
			ChainName:     SedaChainName,
			Version:       config.Images[0].Version,
			ChainConfig:   config,
			NumValidators: &numVals,      // defaults to 2 when unspecified
			NumFullNodes:  &numFullNodes, // defaults to 1 when unspecified
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chains[0] = NewSEDAChain(chains[0].(*cosmos.CosmosChain), logger)

	return chains
}

func BuildAllChains(t *testing.T, chains []ibc.Chain) (*interchaintest.Interchain, context.Context, *client.Client, string) {
	t.Helper()

	ic := interchaintest.NewInterchain()

	sedaChain := chains[0].(*SEDAChain)
	ic = ic.AddChain(sedaChain)

	for _, chain := range chains[1:] {
		ic = ic.AddChain(chain)
	}

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	ctx := context.Background()
	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	})
	require.NoError(t, err)

	return ic, ctx, client, network
}
