package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

//revive:disable-next-line:context-as-argument
func SmartQueryString(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contractAddr, queryMsg string, res interface{}) error {
	t.Helper()
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(queryMsg), &jsonMap); err != nil {
		t.Fatal(err)
	}
	err := chain.QueryContract(ctx, contractAddr, jsonMap, &res)
	return err
}

//revive:disable-next-line:context-as-argument
func SetupAndInstantiateContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname, fileLoc, message string, extraFlags ...string) (codeID, contract string) {
	t.Helper()
	codeID, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}
	needsNoAdminFlag := true

	for _, flag := range extraFlags {
		if flag == "--admin" {
			needsNoAdminFlag = false
		}
	}

	contractAddr, err := chain.InstantiateContract(ctx, keyname, codeID, message, needsNoAdminFlag, extraFlags...)
	if err != nil {
		t.Fatal(err)
	}

	return codeID, contractAddr
}

//revive:disable-next-line:context-as-argument
func MigrateContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname, contractAddr, fileLoc, message string) (codeID, contract string) {
	t.Helper()
	codeID, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}

	// Execute migrate tx
	cmd := []string{
		"seda-chaind", "tx", "wasm", "migrate", contractAddr, codeID, message,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", keyname,
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}

	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}

	return codeID, contractAddr
}

//revive:disable-next-line:context-as-argument
func ExecuteMsgWithAmount(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, message string) {
	t.Helper()
	cmd := buildCmd(chain, user, contractAddr, amount, "", message)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

//revive:disable-next-line:context-as-argument
func ExecuteMsgWithFee(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) {
	t.Helper()
	cmd := buildCmd(chain, user, contractAddr, amount, feeCoin, message)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

//revive:disable-next-line:context-as-argument
func ExecuteMsgWithFeeReturn(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) (*sdk.TxResponse, error) {
	t.Helper()
	// amount is #utoken

	cmd := buildCmd(chain, user, contractAddr, amount, feeCoin, message)
	node := chain.GetNode()
	txHash, err := node.ExecTx(ctx, user.KeyName(), cmd...)
	if err != nil {
		return nil, err
	}

	txRes, err := chain.GetTransaction(txHash)
	return txRes, err
}

func buildCmd(chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) []string {
	cmd := []string{
		"seda-chaind", "tx", "wasm", "execute", contractAddr, message,
		"--node", chain.GetRPCAddress(),
		"--home", chain.HomeDir(),
		"--chain-id", chain.Config().ChainID,
		"--from", user.KeyName(),
		"--gas", "500000",
		"--keyring-dir", chain.HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"-y",
	}

	if amount != "" {
		cmd = append(cmd, "--amount", amount)
	}

	if feeCoin != "" {
		cmd = append(cmd, "--fees", feeCoin)
	}

	return cmd
}
