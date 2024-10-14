package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestSedaGaiaConformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	numVals := 2
	numFullNodes := 1

	gaiaChainSpec := &interchaintest.ChainSpec{
		Name:          "gaia",
		Version:       "v14.1.0",
		NumValidators: &numVals,
		NumFullNodes:  &numFullNodes,
	}
	runConformanceTest(t, gaiaChainSpec, numVals, numFullNodes)
}

func runConformanceTest(t *testing.T, counterpartyChainSpec *interchaintest.ChainSpec, numVals, numFullNodes int) {
	t.Helper()
	/* =================================================== */
	/*                   CHAIN FACTORY                     */
	/* =================================================== */
	cf := interchaintest.NewBuiltinChainFactory(
		zaptest.NewLogger(t),
		[]*interchaintest.ChainSpec{
			{
				Name:          SedaChainName,
				ChainConfig:   GetSEDAConfig(),
				NumValidators: &numVals,
				NumFullNodes:  &numFullNodes,
			},
			counterpartyChainSpec,
		})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	sedaChain, counterpartyChain := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	/* =================================================== */
	/*                  RELAYER FACTORY                    */
	/* =================================================== */
	client, network := interchaintest.DockerSetup(t)

	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		RlyConfig.Type,
		zaptest.NewLogger(t),
		interchaintestrelayer.CustomDockerImage(RlyConfig.Image, RlyConfig.Version, "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	rly := rf.Build(t, client, network)

	/* =================================================== */
	/*                  INTERCHAIN SPAWN                   */
	/* =================================================== */
	const (
		ibcPath = "ibc-path"
	)

	ic := interchaintest.NewInterchain().
		AddChain(sedaChain).
		AddChain(counterpartyChain).
		AddRelayer(rly, RlyConfig.Name).
		AddLink(interchaintest.InterchainLink{
			Chain1:  sedaChain,
			Chain2:  counterpartyChain,
			Relayer: rly,
			Path:    ibcPath,
		})

	ctx := context.Background()
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Perform the conformance test between the two chains
	conformance.TestChainPair(t, ctx, client, network, sedaChain, counterpartyChain, rf, rep, rly, ibcPath)
}
