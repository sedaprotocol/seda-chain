package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testdata"
)

/*
func TestInstantiateWithContractDataResponse(t *testing.T) {
	ctx, keepers := keeper.CreateTestInput(t, false, strings.Join(wasmCapabilities, ","))

	wasmEngineMock := &wasmtesting.MockWasmEngine{
		InstantiateFn: func(codeID wasmvm.Checksum, env wasmvmtypes.Env, info wasmvmtypes.MessageInfo, initMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.Response, uint64, error) {
			return &wasmvmtypes.Response{Ok: &wasmvmtypes.Response{Data: []byte("my-response-data")}}, 0, nil
		},
		// InstantiateFn: func(codeID wasmvm.Checksum, env wasmvmtypes.Env, info wasmvmtypes.MessageInfo, initMsg []byte, store wasmvm.KVStore, goapi wasmvm.GoAPI, querier wasmvm.Querier, gasMeter wasmvm.GasMeter, gasLimit uint64, deserCost wasmvmtypes.UFraction) (*wasmvmtypes.ContractResult, uint64, error) {
		// 	return &wasmvmtypes.ContractResult{Ok: &wasmvmtypes.Response{Data: []byte("my-response-data")}}, 0, nil
		// },
		AnalyzeCodeFn: wasmtesting.WithoutIBCAnalyzeFn,
		StoreCodeFn:   wasmtesting.NoOpStoreCodeFn,
	}

	keeper.StoreRandomContract(t, ctx, keepers, wasmEngineMock)

	example := keeper.StoreExampleContract(t, ctx, keepers, "./test_utils/data_requests.wasm")

	_, _, err := keepers.ContractKeeper.Instantiate(ctx, example.CodeID, example.CreatorAddr, nil, nil, "test", nil)
	require.NoError(t, err)
	// assert.Equal(t, []byte("my-response-data"), data)
}
*/

func TestFetchForTally(t *testing.T) {
	f := initFixture(t)
	ctx := f.Context()

	creator := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Upload code.
	codeID, _, err := f.contractKeeper.Create(ctx, creator, testdata.SedaContractWasm(), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(1), codeID)

	// Instantiate contract.
	initMsg := struct {
		Token string         `json:"token"`
		Owner sdk.AccAddress `json:"owner"`
	}{
		Token: "aseda",
		Owner: creator,
	}
	initMsgBz, err := json.Marshal(initMsg)
	require.NoError(t, err)

	contractAddr, _, err := f.contractKeeper.Instantiate(ctx, codeID, creator, nil, initMsgBz, "DR Contract", sdk.NewCoins())
	require.NoError(t, err)
	require.NotEmpty(t, contractAddr)

	err = f.wasmStorageKeeper.ProxyContractRegistry.Set(ctx, contractAddr.String())
	require.NoError(t, err)

	// Post DR.
	// Q: Is there a way to set up mock contract state for testing?

	// Tally endblock.
	err = f.wasmStorageKeeper.ExecuteTally(ctx)
	require.NoError(t, err)
}
