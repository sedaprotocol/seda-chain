package interchaintest

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

// TestSedaGaiaIBCTransfer spins up a Seda and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Seda->Gaia and then back from Gaia->Seda.
func TestSedaGaiaIBCTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	numVals := 1
	numFullNodes := 1

	// pre configured chain pulled from
	// https://github.com/strangelove-ventures/heighliner
	gaiaChainSpec := &interchaintest.ChainSpec{
		Name:          "gaia",
		Version:       "v14.1.0",
		NumValidators: &numVals,      // defaults to 2 when unspecified
		NumFullNodes:  &numFullNodes, // defaults to 1 when unspecified
	}

	runIBCTransferTest(t, gaiaChainSpec, numVals, numFullNodes)
}

func runIBCTransferTest(t *testing.T, counterpartyChainSpec *interchaintest.ChainSpec, numVals, numFullNodes int) {
	/* =================================================== */
	/*                   CHAIN FACTORY                     */
	/* =================================================== */
	cf := interchaintest.NewBuiltinChainFactory(
		zaptest.NewLogger(t),
		[]*interchaintest.ChainSpec{
			{
				Name:          SedaChainName,
				ChainConfig:   SedaCfg,
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

	/*
	 *	Stake Distribution on Genesis
	 *	  - 2,000,000,000,000 for each validator
	 *	  - 1,000,000,000,000 for each full node
	 *	  - 10,000,000,000 for each faucet on each chain
	 *	  - 1,000,000,000 for relayer
	 */
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

	/* =================================================== */
	/*                  WALLETS & USERS                    */
	/* =================================================== */

	// Create some user accounts on both chains
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), GenesisWalletAmount, sedaChain, counterpartyChain)
	sedaUser, gaiaUser := users[0], users[1]
	sedaUserAddr := sedaUser.FormattedAddress()
	gaiaUserAddr := gaiaUser.FormattedAddress()

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, sedaChain, counterpartyChain)
	require.NoError(t, err)

	// Get original account balances
	sedaOrigBal, err := sedaChain.GetBalance(ctx, sedaUserAddr, sedaChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, GenesisWalletAmount, sedaOrigBal.Int64())

	gaiaOrigBal, err := counterpartyChain.GetBalance(ctx, gaiaUserAddr, counterpartyChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, GenesisWalletAmount, gaiaOrigBal.Int64())

	/* =================================================== */
	/*                  INTERCHAIN TEST                    */
	/* =================================================== */
	var transferAmount = math.NewInt(1_000)

	// Compose an IBC transfer and send from Seda -> Gaia
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   sedaChain.Config().Denom,
		Amount:  transferAmount,
	}

	channel, err := ibc.GetTransferChannel(ctx, rly, eRep, sedaChain.Config().ChainID, counterpartyChain.Config().ChainID)
	require.NoError(t, err)

	sedaHeight, err := sedaChain.Height(ctx)
	require.NoError(t, err)

	transferTx, err := sedaChain.SendIBCTransfer(ctx, channel.ChannelID, sedaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	/*
	 * Starts the relayer on a loop to avoid having to
	 * manually flush packets and ack's
	 */
	err = rly.StartRelayer(ctx, eRep, ibcPath)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := rly.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, sedaChain, sedaHeight, sedaHeight+50, transferTx.Packet)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 10, sedaChain)
	require.NoError(t, err)

	// Get the IBC denom for aseda on Gaia
	sedaTokenDenom := transfertypes.GetPrefixedDenom(channel.Counterparty.PortID, channel.Counterparty.ChannelID, sedaChain.Config().Denom)
	sedaIBCDenom := transfertypes.ParseDenomTrace(sedaTokenDenom).IBCDenom()

	// Assert that the funds are no longer present in user acc on Seda and are in the user acc on Gaia
	sedaUpdateBal, err := sedaChain.GetBalance(ctx, sedaUserAddr, sedaChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, sedaOrigBal.Sub(transferAmount), sedaUpdateBal)

	gaiaUpdateBal, err := counterpartyChain.GetBalance(ctx, gaiaUserAddr, sedaIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, gaiaUpdateBal)

	// Compose an IBC transfer and send from Gaia -> Seda
	transfer = ibc.WalletAmount{
		Address: sedaUserAddr,
		Denom:   sedaIBCDenom,
		Amount:  transferAmount,
	}

	gaiaHeight, err := counterpartyChain.Height(ctx)
	require.NoError(t, err)

	transferTx, err = counterpartyChain.SendIBCTransfer(ctx, channel.Counterparty.ChannelID, gaiaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, counterpartyChain, gaiaHeight, gaiaHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on Seda and not on Gaia
	sedaUpdateBal, err = sedaChain.GetBalance(ctx, sedaUserAddr, sedaChain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, sedaOrigBal, sedaUpdateBal)

	gaiaUpdateBal, err = counterpartyChain.GetBalance(ctx, gaiaUserAddr, sedaIBCDenom)
	require.NoError(t, err)
	require.Equal(t, int64(0), gaiaUpdateBal.Int64())
}
