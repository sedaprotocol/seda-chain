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

// Random tally wasms.
// var mockFetchResponse = []byte(`{"145438bc73d32082ff459e18d1a3db72c50cdaaf908bd90ac9616e265daf4a17":{"commits":{},"dr_binary_id":"fcc85f81d2604dca02fdd1330cc39dc1a6446b8abb16b4f068d8a1d1e2a48877","dr_inputs":"","gas_limit":"20","gas_price":"10","height":5389299128623366229,"id":"145438bc73d32082ff459e18d1a3db72c50cdaaf908bd90ac9616e265daf4a17","memo":"","payback_address":"","replication_factor":2,"reveals":{},"seda_payload":"","tally_binary_id":"3256cbd8d4e68865ebdf3636df7a433fec8999b30a6202a1eb4a9f92363d5550","tally_inputs":"AwMDAwM=","version":"1.0.0"},"573d63b1ea24330e31e18768953699c7ca031e44483cc228a9638010baa8bad0":{"commits":{},"dr_binary_id":"a88516c04de6305f973bf43d7d0112c831eb511a2366df682a335d2aa0dec20b","dr_inputs":"","gas_limit":"20","gas_price":"10","height":7096383960515817382,"id":"573d63b1ea24330e31e18768953699c7ca031e44483cc228a9638010baa8bad0","memo":"","payback_address":"","replication_factor":3,"reveals":{},"seda_payload":"","tally_binary_id":"5d0cef6880aade3af9d59e285bb59ab128327d3ae49f943e31e554ab99b4d21e","tally_inputs":"BgYGBgY=","version":"1.0.0"},"830a37e6d676d4e6ef4458e7d93fa4126caa38efa3df7cfc28c7b0d7997fbe8d":{"commits":{},"dr_binary_id":"a88516c04de6305f973bf43d7d0112c831eb511a2366df682a335d2aa0dec20b","dr_inputs":"","gas_limit":"20","gas_price":"10","height":5515264981634441762,"id":"830a37e6d676d4e6ef4458e7d93fa4126caa38efa3df7cfc28c7b0d7997fbe8d","memo":"","payback_address":"","replication_factor":3,"reveals":{},"seda_payload":"","tally_binary_id":"5d0cef6880aade3af9d59e285bb59ab128327d3ae49f943e31e554ab99b4d21e","tally_inputs":"AgICAgI=","version":"1.0.0"},"b413cf7eb89f35cc44292ed3c6bd7a02bb3fd118c105117ad8e1037e77dd8db8":{"commits":{},"dr_binary_id":"a88516c04de6305f973bf43d7d0112c831eb511a2366df682a335d2aa0dec20b","dr_inputs":"","gas_limit":"20","gas_price":"10","height":14844550059642515049,"id":"b413cf7eb89f35cc44292ed3c6bd7a02bb3fd118c105117ad8e1037e77dd8db8","memo":"","payback_address":"","replication_factor":3,"reveals":{},"seda_payload":"","tally_binary_id":"5d0cef6880aade3af9d59e285bb59ab128327d3ae49f943e31e554ab99b4d21e","tally_inputs":"AwMDAwM=","version":"1.0.0"},"facc77c8dfb3ff645dbc5fb8778ab276841a3e2b4a80c1df5dc28fa516d1816c":{"commits":{},"dr_binary_id":"fcc85f81d2604dca02fdd1330cc39dc1a6446b8abb16b4f068d8a1d1e2a48877","dr_inputs":"","gas_limit":"20","gas_price":"10","height":8215532109948458411,"id":"facc77c8dfb3ff645dbc5fb8778ab276841a3e2b4a80c1df5dc28fa516d1816c","memo":"","payback_address":"","replication_factor":2,"reveals":{},"seda_payload":"","tally_binary_id":"3256cbd8d4e68865ebdf3636df7a433fec8999b30a6202a1eb4a9f92363d5550","tally_inputs":"AwMDAwM=","version":"1.0.0"}}`)

// All requests use sample tally wasm.
var mockFetchResponse2 = []byte(`{"145438bc73d32082ff459e18d1a3db72c50cdaaf908bd90ac9616e265daf4a17":{"commits":{},"dr_binary_id":"fcc85f81d2604dca02fdd1330cc39dc1a6446b8abb16b4f068d8a1d1e2a48877","dr_inputs":"","gas_limit":"20","gas_price":"10","height":5389299128623366229,"id":"145438bc73d32082ff459e18d1a3db72c50cdaaf908bd90ac9616e265daf4a17","memo":"","payback_address":"","replication_factor":2,"reveals":{},"seda_payload":"","tally_binary_id":"2f12d9175337bf340095ee955f8dff5c7baf4cadb0958e63ac4676a6a56fa71e","tally_inputs":"AwMDAwM=","version":"1.0.0"},"573d63b1ea24330e31e18768953699c7ca031e44483cc228a9638010baa8bad0":{"commits":{},"dr_binary_id":"a88516c04de6305f973bf43d7d0112c831eb511a2366df682a335d2aa0dec20b","dr_inputs":"","gas_limit":"20","gas_price":"10","height":7096383960515817382,"id":"573d63b1ea24330e31e18768953699c7ca031e44483cc228a9638010baa8bad0","memo":"","payback_address":"","replication_factor":3,"reveals":{},"seda_payload":"","tally_binary_id":"2f12d9175337bf340095ee955f8dff5c7baf4cadb0958e63ac4676a6a56fa71e","tally_inputs":"BgYGBgY=","version":"1.0.0"},"830a37e6d676d4e6ef4458e7d93fa4126caa38efa3df7cfc28c7b0d7997fbe8d":{"commits":{},"dr_binary_id":"a88516c04de6305f973bf43d7d0112c831eb511a2366df682a335d2aa0dec20b","dr_inputs":"","gas_limit":"20","gas_price":"10","height":5515264981634441762,"id":"830a37e6d676d4e6ef4458e7d93fa4126caa38efa3df7cfc28c7b0d7997fbe8d","memo":"","payback_address":"","replication_factor":3,"reveals":{},"seda_payload":"","tally_binary_id":"2f12d9175337bf340095ee955f8dff5c7baf4cadb0958e63ac4676a6a56fa71e","tally_inputs":"AgICAgI=","version":"1.0.0"},"b413cf7eb89f35cc44292ed3c6bd7a02bb3fd118c105117ad8e1037e77dd8db8":{"commits":{},"dr_binary_id":"a88516c04de6305f973bf43d7d0112c831eb511a2366df682a335d2aa0dec20b","dr_inputs":"","gas_limit":"20","gas_price":"10","height":14844550059642515049,"id":"b413cf7eb89f35cc44292ed3c6bd7a02bb3fd118c105117ad8e1037e77dd8db8","memo":"","payback_address":"","replication_factor":3,"reveals":{},"seda_payload":"","tally_binary_id":"2f12d9175337bf340095ee955f8dff5c7baf4cadb0958e63ac4676a6a56fa71e","tally_inputs":"AwMDAwM=","version":"1.0.0"},"facc77c8dfb3ff645dbc5fb8778ab276841a3e2b4a80c1df5dc28fa516d1816c":{"commits":{},"dr_binary_id":"fcc85f81d2604dca02fdd1330cc39dc1a6446b8abb16b4f068d8a1d1e2a48877","dr_inputs":"","gas_limit":"20","gas_price":"10","height":8215532109948458411,"id":"facc77c8dfb3ff645dbc5fb8778ab276841a3e2b4a80c1df5dc28fa516d1816c","memo":"","payback_address":"","replication_factor":2,"reveals":{},"seda_payload":"","tally_binary_id":"2f12d9175337bf340095ee955f8dff5c7baf4cadb0958e63ac4676a6a56fa71e","tally_inputs":"AwMDAwM=","version":"1.0.0"}}`)

var drWasmByteArray = []byte("82a9dda829eb7f8ffe9fbe49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e5082")

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

	tallyWasm := types.NewWasm(testdata.SampleTallyWasm(), types.WasmTypeDataRequest, ctx.BlockTime(), ctx.BlockHeight(), 100)
	err = k.DataRequestWasm.Set(ctx, tallyWasm.Hash, tallyWasm)
	require.NoError(t, err)

	err = k.ExecuteTally(ctx)
	require.NoError(t, err)
}
