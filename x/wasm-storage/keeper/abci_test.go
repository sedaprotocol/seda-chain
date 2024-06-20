package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/sedaprotocol/seda-chain/app/params"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

// All requests use the sample tally wasm and filter type none.
var mockFetchResponse = []byte(`[{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":1661661742461173125,"id":"fba5314c57e52da7d1a2245d18c670fde1cb8c237062d2a1be83f449ace0932e","memo":"","payback_address":"","replication_factor":3,"reveals":{"1b85dfb9420e6757630a0db2280fa1787ec8c1e419a6aca76dbbfe8ef6e17521":{"exit_code":0,"gas_used":"10","reveal":"Ng==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"1dae290cd880b79d21079d89aee3460cf8a7d445fb35cade70cf8aa96924441c":{"exit_code":0,"gas_used":"10","reveal":"LQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"},"421e735518ef77fc1209a9d3585cdf096669b52ea68549e2ce048d4919b4c8c0":{"exit_code":0,"gas_used":"10","reveal":"DQ==","salt":"05952214b2ba3549a8d627c57d2d0dd1b0a2ce65c46e3b2f25c273464be8ba5f"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"},{"commits":{},"dr_binary_id":"9471d36add157cd7eaa32a42b5ddd091d5d5d396bf9ad67938a4fc40209df6cf","dr_inputs":"","gas_limit":"20","gas_price":"10","height":9859593541233596221,"id":"d4e40f45fbf529134926acf529baeb6d4f37b5c380d7ab6b934833e7c00d725f","memo":"","payback_address":"","replication_factor":1,"reveals":{"c9a4c8f1e70a0059a88b4768a920e41c95c587b8387ea3286d8fa4ee3b68b038":{"exit_code":0,"gas_used":"10","reveal":"Yw==","salt":"f837455a930a66464f1c50586dc745a6b14ea807727c6069acac24c9558b6dbf"}},"seda_payload":"","tally_binary_id":"8ade60039246740faa80bf424fc29e79fe13b32087043e213e7bc36620111f6b","tally_inputs":"AAEBAQE=","version":"1.0.0"}]`)

func TestExecuteTally(t *testing.T) {
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: params.Bech32PrefixAccPub,
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: params.Bech32PrefixValAddr,
			},
		},
	})
	require.NoError(t, err)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := sdktestutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})

	ctrl := gomock.NewController(t)
	accKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	contractOpsKeeper := testutil.NewMockContractOpsKeeper(ctrl)
	viewKeeper := testutil.NewMockViewKeeper(ctrl)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		accKeeper,
		bankKeeper,
		contractOpsKeeper,
		viewKeeper,
	)

	// Add random address to contract registry and mock contract query result.
	err = k.ProxyContractRegistry.Set(ctx, "seda1ucv5709wlf9jn84ynyjzyzeavwvurmdyxat26l")
	require.NoError(t, err)
	viewKeeper.EXPECT().QuerySmart(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockFetchResponse, nil)

	// Store the sample tally that will be used.
	tallyWasm := types.NewWasm(testdata.SampleTallyWasm(), types.WasmTypeDataRequest, ctx.BlockTime(), ctx.BlockHeight(), 100)
	err = k.DataRequestWasm.Set(ctx, tallyWasm.Hash, tallyWasm)
	require.NoError(t, err)

	err = k.ExecuteTally(ctx)
	require.NoError(t, err)
}
