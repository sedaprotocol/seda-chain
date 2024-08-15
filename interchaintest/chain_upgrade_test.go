package interchaintest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/sedaprotocol/seda-chain/interchaintest/conformance"
	"github.com/sedaprotocol/seda-chain/interchaintest/helpers"
)

const (
	upgradeName        = "v1"
	initialVersion     = "old"
	upgradeVersion     = "new"
	upgradeRepo        = "sedad-e2e"
	haltHeightDelta    = uint64(10) // # of blocks after which to submit upgrade proposal
	blocksAfterUpgrade = uint64(10) // # of blocks to wait after upgrade is applied
)

var (
	numVals, numFullNodes = 4, 0

	// current chain version we are upgrading from
	baseChain = ibc.DockerImage{
		Repository: "sedad-e2e", // to be replaced by sedaRepo once we have Docker images setup
		Version:    initialVersion,
		UidGid:     "1025:1025",
	}
)

func TestChainUpgrade(t *testing.T) {
	BasicUpgradeTest(t, upgradeVersion, upgradeRepo, upgradeName)
}

func BasicUpgradeTest(t *testing.T, upgradeVersion, upgradeRepo, upgradeName string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(SedaChainName, upgradeVersion, upgradeRepo, upgradeName)

	cfg := SedaCfg
	cfg.ModifyGenesis = cosmos.ModifyGenesis(getTestGenesis())
	cfg.Images = []ibc.DockerImage{baseChain}

	chains := CreateChainsWithCustomConfig(t, numVals, numFullNodes, cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ic, ctx, client, _ := BuildAllChains(t, chains)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	// fund cghain user
	chainUser := fundChainUser(t, ctx, chain)

	// upload and execute a contract before the upgrade
	beforeContract := conformance.Basic(t, ctx, chain, chainUser)

	// submit upgrade proposal
	haltHeight, proposalID := SubmitUpgradeProposal(t, ctx, chain, chainUser, upgradeName)

	currentHeight, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// vote
	ValidatorVoting(t, ctx, chain, proposalID, currentHeight, haltHeight)

	// upgrade
	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, upgradeVersion)

	// load beforeContract from disk and confirm we can still execute against it post upgrade
	helpers.ExecuteMsgWithFee(t, ctx, chain, chainUser, beforeContract, "", "1000000000000000000"+chain.Config().Denom, `{"increment":{}}`)

	// test conformance after upgrade
	conformance.CosmWasm(t, ctx, chain, chainUser)
}

//revive:disable-next-line:context-as-argument
func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight uint64, upgradeRepo, upgradeVersion string) {
	t.Helper()
	stopNodes(t, ctx, chain)
	upgradeNodes(t, ctx, chain, client, upgradeRepo, upgradeVersion)
	startNodes(t, ctx, chain)
	waitForBlocks(t, ctx, chain)
	checkHeight(t, ctx, chain, haltHeight)
}

//revive:disable-next-line:context-as-argument
func fundChainUser(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) ibc.Wallet {
	t.Helper()
	userFunds := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	return users[0]
}

func getTestGenesis() []cosmos.GenesisKV {
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
	}
}

//revive:disable-next-line:context-as-argument
func stopNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) {
	t.Helper()
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(ctx)
	if err != nil {
		t.Fatalf("error stopping node(s): %v", err)
	}
}

//revive:disable-next-line:context-as-argument
func upgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, upgradeRepo, upgradeVersion string) {
	t.Helper()
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeVersion)
}

//revive:disable-next-line:context-as-argument
func startNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) {
	t.Helper()
	t.Log("starting node(s)")
	err := chain.StartAllNodes(ctx)
	if err != nil {
		t.Fatalf("error starting upgraded node(s): %v", err)
	}
}

//revive:disable-next-line:context-as-argument
func waitForBlocks(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) {
	t.Helper()
	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err := testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	if err != nil {
		t.Fatalf("chain did not produce blocks after upgrade: %v", err)
	}
}

//revive:disable-next-line:context-as-argument
func checkHeight(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, haltHeight uint64) {
	t.Helper()
	height, err := chain.Height(ctx)
	if err != nil {
		t.Fatalf("error fetching height after upgrade: %v", err)
	}

	if height < haltHeight+blocksAfterUpgrade {
		t.Fatalf("height did not increment enough after upgrade")
	}
}

// see https://github.com/golangci/golangci-lint/issues/741
// Ignore SA4009 throughout the file due to specific use case.
//
//nolint:staticcheck // SA4009 disable staticcheck since there is a bug where you can't ignore a single lint check:
//revive:disable-next-line:context-as-argument
func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, currentHeight, haltHeight uint64) {
	t.Helper()

	err := chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	/*
	 * PollForProposalStatus gets stuck and ends up blocking execution,
	 * preventing nodes from upgrading; Disabling for now
	 */

	// _, err = cosmos.PollForProposalStatus(ctx, chain, currentHeight, currentHeight+haltHeightDelta, proposalID, cosmos.ProposalStatusPassed)
	// require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*100)
	defer timeoutCtxCancel()

	currentHeight, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-currentHeight), chain)

	currentHeight, err = chain.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, currentHeight, "height is not equal to halt height")
}

//revive:disable-next-line:context-as-argument
func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string) (uint64, string) {
	t.Helper()
	currentHeight, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := currentHeight + haltHeightDelta

	upgradeMsg := &upgradetypes.MsgSoftwareUpgrade{
		Authority: "seda10d07y265gmmuvt4z0w9aw880jnsr700jvvla4j", // gov module account; sedad q auth module-account gov
		Plan: upgradetypes.Plan{
			Name:   upgradeName,
			Height: int64(haltHeight),
		},
	}

	proposal, err := chain.BuildProposal([]cosmos.ProtoMessage{upgradeMsg},
		"Chain Upgrade 1", // title
		"Summary desc",    // summary
		"ipfs://CID",      // metadata
		fmt.Sprintf(`500000000%s`, chain.Config().Denom), // deposit string
		string(user.Address()),                           // proposer address
		false,                                            // expedited
	)
	require.NoError(t, err, "error building proposal")

	txProp, err := chain.SubmitProposal(ctx, user.KeyName(), proposal)
	t.Log("txProp", txProp)
	require.NoError(t, err, "error submitting proposal")

	return haltHeight, txProp.ProposalID
}
