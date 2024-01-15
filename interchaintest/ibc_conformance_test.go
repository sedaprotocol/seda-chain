package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGaiaConformance(t *testing.T) {
	numVals := 2
	numFullNodes := 1

	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	gaiaChainSpec := &interchaintest.ChainSpec{
		Name:          "gaia",
		Version:       "v14.1.0",
		NumValidators: &numVals,      // defaults to 2 when unspecified
		NumFullNodes:  &numFullNodes, // defaults to 1 when unspecified
	}
	runConformanceTest(t, gaiaChainSpec, numVals, numFullNodes)
}

func runConformanceTest(t *testing.T, gaiaChainSpec *interchaintest.ChainSpec, numVals, numFullNodes int) {
	/* =================================================== */
	/*                   CHAIN FACTORY                     */
	/* =================================================== */
	chainFactory := interchaintest.NewBuiltinChainFactory(
		zaptest.NewLogger(t),
		[]*interchaintest.ChainSpec{
			{
				Name:          "seda",
				ChainConfig:   SedaCfg,
				NumValidators: &numVals,      // defaults to 2 when unspecified
				NumFullNodes:  &numFullNodes, // defaults to 1 when unspecified
			},
			gaiaChainSpec,
		})

	// Get chains from the chain factory
	chains, err := chainFactory.Chains(t.Name())
	require.NoError(t, err)
	sedaChain, counterpartyChain := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	/* =================================================== */
	/*                  RELAYER FACTORY                    */
	/* =================================================== */
	const (
		relayerType = ibc.CosmosRly
		relayerName = "relay"
	)
	client, network := interchaintest.DockerSetup(t)

	// Get a relayer instance
	rf := interchaintest.NewBuiltinRelayerFactory(
		relayerType,
		zaptest.NewLogger(t),
		interchaintestrelayer.CustomDockerImage(RelayerImage, RelayerVersion, "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	relayer := rf.Build(t, client, network)

	/* =================================================== */
	/*                  INTERCHAIN SPAWN                   */
	/* =================================================== */
	const (
		ibcPath = "ibc-path"
	)

	ic := interchaintest.NewInterchain().
		AddChain(sedaChain).
		AddChain(counterpartyChain).
		AddRelayer(relayer, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  sedaChain,
			Chain2:  counterpartyChain,
			Relayer: relayer,
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
	conformance.TestChainPair(t, ctx, client, network, sedaChain, counterpartyChain, rf, rep, relayer, ibcPath)
}
