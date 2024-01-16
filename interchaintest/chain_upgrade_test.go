package interchaintest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/sedaprotocol/seda-chain/interchaintest/conformance"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	cosmosproto "github.com/cosmos/gogoproto/proto"
)

const (
	upgradeName        = "test_upgrade"
	upgradeVersion     = "v1.1.0"
	upgradeRepo        = "test_repo"
	haltHeightDelta    = uint64(12) // propose upgrade x blocks after current height
	blocksAfterUpgrade = uint64(6)  // blocks to wait after upgrade before checking height
)

var (
	numVals, numFullNodes = 4, 0

	// current chain version we are upgrading from
	baseChain = ibc.DockerImage{
		Repository: "sedaprotocol/seda-chaind-e2e", // to be replaced by sedaRepo once we have Docker images setup
		Version:    "upgrade",
		UidGid:     "1025:1025",
	}
)

func TestChainUpgrade(t *testing.T) {
	BasicUpgradeTest(t, SedaChainName, upgradeVersion, upgradeRepo, upgradeName)
}

func BasicUpgradeTest(t *testing.T, SedaChainName, upgradeBranchVersion, upgradeRepo, upgradeName string) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	t.Log(SedaChainName, upgradeBranchVersion, upgradeRepo, upgradeName)

	previousVersionGenesis := []cosmos.GenesisKV{
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

	cfg := SedaCfg
	cfg.ModifyGenesis = cosmos.ModifyGenesis(previousVersionGenesis)
	cfg.Images = []ibc.DockerImage{baseChain}

	chains := CreateChainsWithCustomConfig(t, numVals, numFullNodes, cfg)
	chain := chains[0].(*cosmos.CosmosChain)

	ic, ctx, client, _ := BuildAll(t, chains)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chain)
	chainUser := users[0]

	// submit upgrade proposal
	currentHeight, err := chain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := currentHeight + haltHeightDelta
	proposalID := SubmitUpgradeProposal(t, ctx, chain, chainUser, upgradeName, haltHeight)

	// vote
	ValidatorVoting(t, ctx, chain, proposalID, currentHeight, haltHeight)

	// upgrade
	UpgradeNodes(t, ctx, chain, client, haltHeight, upgradeRepo, upgradeBranchVersion)

	// test conformance after upgrade
	conformance.ConformanceCosmWasm(t, ctx, chain, chainUser)
}

func UpgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, haltHeight uint64, upgradeRepo, upgradeBranchVersion string) {
	stopNodes(t, ctx, chain)
	upgradeNodes(t, ctx, chain, client, upgradeRepo, upgradeBranchVersion)
	startNodes(t, ctx, chain)
	waitForBlocks(t, ctx, chain)
	checkHeight(t, ctx, chain, haltHeight)
}

func stopNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) {
	t.Log("stopping node(s)")
	err := chain.StopAllNodes(ctx)
	if err != nil {
		t.Fatalf("error stopping node(s): %v", err)
	}
}

func upgradeNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, client *client.Client, upgradeRepo, upgradeBranchVersion string) {
	t.Log("upgrading node(s)")
	chain.UpgradeVersion(ctx, client, upgradeRepo, upgradeBranchVersion)
}

func startNodes(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) {
	t.Log("starting node(s)")
	err := chain.StartAllNodes(ctx)
	if err != nil {
		t.Fatalf("error starting upgraded node(s): %v", err)
	}
}

func waitForBlocks(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain) {
	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*60)
	defer timeoutCtxCancel()

	err := testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), chain)
	if err != nil {
		t.Fatalf("chain did not produce blocks after upgrade: %v", err)
	}
}

func checkHeight(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, haltHeight uint64) {
	height, err := chain.Height(ctx)
	if err != nil {
		t.Fatalf("error fetching height after upgrade: %v", err)
	}

	if height < haltHeight+blocksAfterUpgrade {
		t.Fatalf("height did not increment enough after upgrade")
	}
}

func ValidatorVoting(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, proposalID string, currentHeight uint64, haltHeight uint64) {
	err := chain.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, chain, currentHeight, currentHeight+haltHeightDelta, proposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
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

func SubmitUpgradeProposal(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, upgradeName string, haltHeight uint64) string {
	upgradeMsg := []cosmosproto.Message{
		&upgradetypes.MsgSoftwareUpgrade{
			Authority: "to-do", // gov module account
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
			},
		},
	}

	proposal, err := chain.BuildProposal(upgradeMsg,
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

	return txProp.ProposalID
}
