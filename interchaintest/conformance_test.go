package interchaintest

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"go.uber.org/zap/zaptest"
)

/* TestConformance tests the followings:
 *   - Client, channel, and connection creation
 *   - Messages are properly relayed and acknowledged
 *   - Packets are being properly timed out
 */
func TestConformance(t *testing.T) {

	/* =================================================== */
	/*                   CHAIN FACTORY                     */
	/* =================================================== */
	numOfValidators := 1
	numOfFullNodes := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "seda",
			ChainConfig:   SedaCfg,
			NumValidators: &numOfValidators, // defaults to 2 when unspecified
			NumFullNodes:  &numOfFullNodes,  // defaults to 1 when unspecified
		},
		// pre configured chain pulled from
		// https://github.com/strangelove-ventures/heighliner
		{
			Name:          "gaia",
			Version:       "v14.1.0",
			NumValidators: &numOfValidators,
			NumFullNodes:  &numOfFullNodes,
		},
	})

	/* =================================================== */
	/*                  RELAYER FACTORY                    */
	/* =================================================== */
	rlyFactory := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
	)

	hermesFactory := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		zaptest.NewLogger(t),
	)

	ctx := context.Background()

	// don't collect test reports
	rep := testreporter.NewNopReporter()

	/*
	 * Test against both chains, ensuring that they both have basic IBC capabilities
	 * properly implemented and work with both the Go relayer and Hermes
	 */
	conformance.Test(t, ctx, []interchaintest.ChainFactory{cf}, []interchaintest.RelayerFactory{rlyFactory, hermesFactory}, rep)
}
