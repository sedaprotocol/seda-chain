package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	wasmstorage "github.com/sedaprotocol/seda-chain/x/wasm-storage"
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/keeper"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

func setupKeeper(t *testing.T) (*keeper.Keeper, moduletestutil.TestEncodingConfig, sdk.Context) {
	t.Helper()
	key := sdk.NewKVStoreKey(wasmstoragetypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig(wasmstorage.AppModuleBasic{})
	wasmstoragetypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	msr := baseapp.NewMsgServiceRouter()

	wasmStorageKeeper := keeper.NewKeeper(encCfg.Codec, key, authtypes.NewModuleAddress("gov").String())

	msr.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	wasmstoragetypes.RegisterMsgServer(msr, keeper.NewMsgServerImpl(*wasmStorageKeeper))
	return wasmStorageKeeper, encCfg, ctx
}
