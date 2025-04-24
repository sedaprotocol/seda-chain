package interchaintest

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

const (
	stateSyncSnapshotInterval = 10
	delta                     = stateSyncSnapshotInterval * 2
)

func TestStateSync(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	numOfValidators := 1
	numOfFullNodes := 1

	// snapshot-interval must be a multiple of pruning-keep-every
	configFileOverrides := modifyAppToml(stateSyncSnapshotInterval)

	chains := CreateChains(t, numOfValidators, numOfFullNodes, configFileOverrides)
	chain := chains[0].(*SEDAChain)

	_, ctx, _, _ := BuildAllChains(t, chains)

	// wait for blocks so that nodes have a few state sync snapshot available
	require.NoError(t, testutil.WaitForBlocks(ctx, delta, chain))

	latestHeight, err := chain.Height(ctx)
	require.NoError(t, err, "failed to fetch latest chain height")

	trustHeight := int64(latestHeight) - stateSyncSnapshotInterval

	firstFullNode := chain.FullNodes[0]

	// fetch block hash for trusted height from full node
	blockRes, err := firstFullNode.Client.Block(ctx, &trustHeight)
	require.NoError(t, err, "failed to fetch trusted block")
	trustHash := hex.EncodeToString(blockRes.BlockID.Hash)

	configFileOverrides = modifyConfigToml(configFileOverrides, trustHash, trustHeight, firstFullNode.HostName())

	// state sync a new node
	require.NoError(t, chain.AddFullNodes(ctx, configFileOverrides, 1))

	// wait for the new node to catch up with the chain
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	require.NoError(t, testutil.WaitForInSync(ctx, chain, chain.FullNodes[len(chain.FullNodes)-1]))
}

// snapshot settings & overrides in config.toml
func modifyConfigToml(configFileOverrides map[string]any, trustHash string, trustHeight int64, hostName string) map[string]any {
	configTomlOverrides := make(testutil.Toml)

	stateSync := make(testutil.Toml)
	stateSync["trust_hash"] = trustHash
	stateSync["trust_height"] = trustHeight
	stateSync["rpc_servers"] = fmt.Sprintf("tcp://%s:26657,tcp://%s:26657", hostName, hostName)
	configTomlOverrides["statesync"] = stateSync
	configFileOverrides["config/config.toml"] = configTomlOverrides

	return configFileOverrides
}

func modifyAppToml(snapshotInterval int) map[string]any {
	appTomlOverrides := make(testutil.Toml)

	// snapshot settings & overrides in app.toml
	stateSync := make(testutil.Toml)
	stateSync["snapshot-interval"] = snapshotInterval

	// snapshot-interval must be a multiple of pruning-keep-every
	appTomlOverrides["state-sync"] = stateSync
	appTomlOverrides["pruning"] = "custom"
	appTomlOverrides["pruning-keep-recent"] = snapshotInterval
	appTomlOverrides["pruning-keep-every"] = snapshotInterval
	appTomlOverrides["pruning-interval"] = snapshotInterval

	sedaAppTomlOverrides := getSEDAAppTomlOverrides()
	for k, v := range sedaAppTomlOverrides {
		appTomlOverrides[k] = v
	}

	configFileOverrides := make(map[string]any)
	configFileOverrides["config/app.toml"] = appTomlOverrides
	return configFileOverrides
}
