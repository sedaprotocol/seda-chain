package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var (
	coinType = "118"
	denom    = "aseda"

	// TO-DO
	dockerImage = ibc.DockerImage{
		Repository: "sedaprotocol/seda-chaind-e2e", // FOR LOCAL IMAGE USE: Docker Image Name
		Version:    "latest",                       // FOR LOCAL IMAGE USE: Docker Image Tag
		UidGid:     "1025:1025",
	}

	SedaCfg = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "seda-local",
		ChainID:             "seda-local-1",
		Images:              []ibc.DockerImage{dockerImage},
		Bin:                 "seda-chaind",
		Bech32Prefix:        "seda",
		Denom:               denom,
		CoinType:            coinType,
		GasPrices:           fmt.Sprintf("0%s", denom),
		GasAdjustment:       2.0,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		SkipGenTx:           false,
		PreGenesis:          nil,
		ModifyGenesis:       nil,
		ConfigFileOverrides: nil,
	}

	RelayerImage   = "ghcr.io/cosmos/relayer"
	RelayerVersion = "main"

	GenesisWalletAmount = int64(10_000_000)
)

// CreateChains generates this branch's chain (ex: from the commit)
func CreateChains(t *testing.T, numVals, numFullNodes int) []ibc.Chain {
	cfg := SedaCfg
	cfg.Images = []ibc.DockerImage{dockerImage}
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "seda",
			ChainConfig:   SedaCfg,
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

func BuildAll(t *testing.T, chains []ibc.Chain) (*interchaintest.Interchain, context.Context, *client.Client, string) {
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
