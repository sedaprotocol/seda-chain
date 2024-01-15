package helpers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func SmartQueryString(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, contractAddr, queryMsg string, res interface{}) error {
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(queryMsg), &jsonMap); err != nil {
		t.Fatal(err)
	}
	err := chain.QueryContract(ctx, contractAddr, jsonMap, &res)
	return err
}

func SetupAndInstantiateContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname string, fileLoc string, message string, extraFlags ...string) (codeId, contract string) {
	codeId, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}
	needsNoAdminFlag := true

	for _, flag := range extraFlags {
		if flag == "--admin" {
			needsNoAdminFlag = false
		}
	}

	contractAddr, err := chain.InstantiateContract(ctx, keyname, codeId, message, needsNoAdminFlag, extraFlags...)
	if err != nil {
		t.Fatal(err)
	}

	return codeId, contractAddr
}

func MigrateContract(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, keyname string, contractAddr string, fileLoc string, message string) (codeId, contract string) {
	codeId, err := chain.StoreContract(ctx, keyname, fileLoc)
	if err != nil {
		t.Fatal(err)
	}

	// Execute migrate tx
	cmd := []string{
		"seda-chaind", "tx", "wasm", "migrate", contractAddr, codeId, message,
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

	return codeId, contractAddr
}

func ExecuteMsgWithAmount(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, message string) {
	cmd := buildCmd(chain, user, contractAddr, amount, "", message)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

func ExecuteMsgWithFee(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) {
	cmd := buildCmd(chain, user, contractAddr, amount, feeCoin, message)
	stdout, _, err := chain.Exec(ctx, cmd, nil)
	require.NoError(t, err)

	t.Log(string(stdout))

	if err := testutil.WaitForBlocks(ctx, 2, chain); err != nil {
		t.Fatal(err)
	}
}

func ExecuteMsgWithFeeReturn(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet, contractAddr, amount, feeCoin, message string) (*sdk.TxResponse, error) {
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
