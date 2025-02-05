package v1

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sedaprotocol/seda-chain/app/keepers"
	"github.com/sedaprotocol/seda-chain/app/upgrades"
	batchingtypes "github.com/sedaprotocol/seda-chain/x/batching/types"
	dataproxytypes "github.com/sedaprotocol/seda-chain/x/data-proxy/types"
	pubkeytypes "github.com/sedaprotocol/seda-chain/x/pubkey/types"
	tallytypes "github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	UpgradeName = "v1.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added:   []string{wasmstoragetypes.StoreKey, batchingtypes.StoreKey, tallytypes.StoreKey, dataproxytypes.StoreKey, pubkeytypes.StoreKey},
		Deleted: []string{},
	},
}

func CreateUpgradeHandler(
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Add additional upgrade logic when needed

		/*
		 * migrations are run in module name alphabetical
		 * ascending order, except x/auth which is run last
		 */
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
