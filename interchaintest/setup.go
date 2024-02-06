package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/docker/docker/client"
	"github.com/sedaprotocol/seda-chain/interchaintest/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	ibclocalhost "github.com/cosmos/ibc-go/v8/modules/light-clients/09-localhost"
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

	SedaCfg = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                SedaChainName,
		ChainID:             "seda-local-1",
		Images:              []ibc.DockerImage{dockerImage},
		Bin:                 "sedad",
		Bech32Prefix:        "seda",
		Denom:               SedaDenom,
		CoinType:            coinType,
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
	VotingPeriod     = "15s"
	MaxDepositPeriod = "10s"

	/* =================================================== */
	/*                    WALLET CONFIG                    */
	/* =================================================== */
	GenesisWalletAmount = math.NewInt(10_000_000_000)
)

// sedaEncoding registers the Juno specific module codecs so that the associated types and msgs
// will be supported when writing to the blocksdb sqlite database.
func sedaEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	ibclocalhost.RegisterInterfaces(cfg.InterfaceRegistry)
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

// CreateChains generates this branch's chain (ex: from the commit)
func CreateChains(t *testing.T, numVals, numFullNodes int) []ibc.Chain {
	t.Helper()
	cfg := SedaCfg
	cfg.Images = []ibc.DockerImage{dockerImage}
	return CreateChainsWithCustomConfig(t, numVals, numFullNodes, cfg)
}

func CreateChainsWithCustomConfig(t *testing.T, numVals, numFullNodes int, config ibc.ChainConfig) []ibc.Chain {
	t.Helper()
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
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

	// chain := chains[0].(*cosmos.CosmosChain)
	return chains
}

func BuildAllChains(t *testing.T, chains []ibc.Chain) (*interchaintest.Interchain, context.Context, *client.Client, string) {
	t.Helper()
	ic := interchaintest.NewInterchain()

	for _, chain := range chains {
		ic = ic.AddChain(chain)
	}

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err := ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	})
	require.NoError(t, err)

	return ic, ctx, client, network
}
