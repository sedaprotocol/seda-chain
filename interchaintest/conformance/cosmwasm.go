package conformance

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/sedaprotocol/seda-chain/interchaintest/helpers"
	"github.com/sedaprotocol/seda-chain/interchaintest/types"
)

// CosmWasm validates that store, instantiate, execute, and query work on a CosmWasm contract.
//
//revive:disable-next-line:context-as-argument
func CosmWasm(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) {
	t.Helper()
	Basic(t, ctx, chain, user)
}

//revive:disable-next-line:context-as-argument
func Basic(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, user ibc.Wallet) (contractAddr string) {
	t.Helper()
	_, contractAddr = helpers.SetupAndInstantiateContract(t, ctx, chain, user.KeyName(), "contracts/cw_template.wasm", `{"count":0}`)
	helpers.ExecuteMsgWithFee(t, ctx, chain, user, contractAddr, "", "10000"+chain.Config().Denom, `{"increment":{}}`)

	var res types.GetCountResponse
	err := helpers.SmartQueryString(t, ctx, chain, contractAddr, `{"get_count":{}}`, &res)
	require.NoError(t, err)

	require.Equal(t, int64(1), res.Data.Count)

	return contractAddr
}
