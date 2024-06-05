package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/core/header"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/tx/signing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

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
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

var drWasmByteArray = []byte("82a9dda829eb7f8ffe9fbe49e45d47d2dad9664fbb7adf72492e3c81ebd3e29134d9bc12212bf83c6840f10e8246b9db54a4859b7ccd0123d86e5872c1e5082")
var mockFetchResponse = []byte(`{"9682bc70788e54e9a7b84098b1227cc54c4fe3f1d6ce61e4497b9b09a9dc4447":{"request":{"commits":{},"dr_binary_id":[50,155,207,139,25,73,34,249,14,43,182,40,178,107,229,147,10,91,141,80,235,37,228,233,145,175,155,29,118,0,41,48],"dr_inputs":[],"gas_limit":"20","gas_price":"10","id":[150,130,188,112,120,142,84,233,167,184,64,152,177,34,124,197,76,79,227,241,214,206,97,228,73,123,155,9,169,220,68,71],"memo":[],"payback_address":[],"replication_factor":3,"reveals":{},"seda_payload":[],"tally_binary_id":[18,248,107,209,66,130,224,204,172,84,79,28,42,22,233,148,19,156,229,80,85,192,236,154,175,241,232,225,158,14,63,253],"tally_inputs":[6,6,6,6,6],"version":"1.0.0"}},"f92d5fe38b8c9189d98f90c70a526c8e683496aebebeec06e61b7aeabbb843e9":{"request":{"commits":{},"dr_binary_id":[50,155,207,139,25,73,34,249,14,43,182,40,178,107,229,147,10,91,141,80,235,37,228,233,145,175,155,29,118,0,41,48],"dr_inputs":[],"gas_limit":"20","gas_price":"10","id":[249,45,95,227,139,140,145,137,217,143,144,199,10,82,108,142,104,52,150,174,190,190,236,6,230,27,122,234,187,184,67,233],"memo":[],"payback_address":[],"replication_factor":3,"reveals":{},"seda_payload":[],"tally_binary_id":[18,248,107,209,66,130,224,204,172,84,79,28,42,22,233,148,19,156,229,80,85,192,236,154,175,241,232,225,158,14,63,253],"tally_inputs":[7,7,7,7,7],"version":"1.0.0"}},"fb1ce58d12bd5a290225d911bd2ff87f7ac81269578ac8d3ce655c1ee300a99d":{"request":{"commits":{},"dr_binary_id":[50,155,207,139,25,73,34,249,14,43,182,40,178,107,229,147,10,91,141,80,235,37,228,233,145,175,155,29,118,0,41,48],"dr_inputs":[],"gas_limit":"20","gas_price":"10","id":[251,28,229,141,18,189,90,41,2,37,217,17,189,47,248,127,122,200,18,105,87,138,200,211,206,101,92,30,227,0,169,157],"memo":[],"payback_address":[],"replication_factor":3,"reveals":{},"seda_payload":[],"tally_binary_id":[18,248,107,209,66,130,224,204,172,84,79,28,42,22,233,148,19,156,229,80,85,192,236,154,175,241,232,225,158,14,63,253],"tally_inputs":[9,9,9,9,9],"version":"1.0.0"}}}`)

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

	err = k.ExecuteTally(ctx)
	require.NoError(t, err)
}
