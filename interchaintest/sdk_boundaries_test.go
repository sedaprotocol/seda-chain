package interchaintest

import (
	"context"
	"testing"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/rly"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type boundarySpecs struct {
	name           string
	chainSpecs     []*interchaintest.ChainSpec
	relayerVersion string
}

func TestSDKBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	tests := []boundarySpecs{
		{
			name: "SEDA Chain <-> SDK 47",
			chainSpecs: []*interchaintest.ChainSpec{
				{
					Name:          SedaChainName,
					ChainConfig:   GetSEDAConfig(),
					NumValidators: &numVals,
					NumFullNodes:  &numFullNodes,
				},
				{
					Name:      "ibc-go-simd",
					ChainName: "simd-47",
					Version:   "v7.2.0", // sdk 0.47.3
				},
			},
			relayerVersion: "colin-event-fix",
		},
	}

	for _, tt := range tests {
		tt := tt
		testname := tt.name
		t.Run(testname, func(t *testing.T) {
			t.Parallel()

			logger := zaptest.NewLogger(t)
			cf := interchaintest.NewBuiltinChainFactory(logger, tt.chainSpecs)

			chains, err := cf.Chains(t.Name())
			require.NoError(t, err)

			sedaChain := NewSEDAChain(chains[0].(*cosmos.CosmosChain), logger)
			counterpartyChain := chains[1].(*cosmos.CosmosChain)

			client, network := interchaintest.DockerSetup(t)

			const (
				path        = "ibc-path"
				relayerName = "relayer"
			)

			// Get a relayer instance
			rf := interchaintest.NewBuiltinRelayerFactory(
				ibc.CosmosRly,
				zaptest.NewLogger(t),
				relayer.CustomDockerImage(
					rly.DefaultContainerImage,
					tt.relayerVersion,
					rly.RlyDefaultUidGid,
				),
			)

			r := rf.Build(t, client, network)

			ic := interchaintest.NewInterchain().
				AddChain(sedaChain).
				AddChain(counterpartyChain).
				AddRelayer(r, relayerName).
				AddLink(interchaintest.InterchainLink{
					Chain1:  sedaChain,
					Chain2:  counterpartyChain,
					Relayer: r,
					Path:    path,
				})

			ctx := context.Background()

			rep := testreporter.NewNopReporter()

			require.NoError(t, ic.Build(ctx, rep.RelayerExecReporter(t), interchaintest.InterchainBuildOptions{
				TestName:          t.Name(),
				Client:            client,
				NetworkID:         network,
				BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
				SkipPathCreation:  false,
			}))
			t.Cleanup(func() {
				_ = ic.Close()
			})

			// test IBC conformance
			conformance.TestChainPair(t, ctx, client, network, sedaChain, counterpartyChain, rf, rep, r, path)
		})
	}
}
