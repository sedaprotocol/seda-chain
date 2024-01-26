package v1

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sedaprotocol/seda-chain/app/keepers"
	"github.com/sedaprotocol/seda-chain/app/upgrades"
)

const (
	UpgradeName = "v1"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: Createv1UpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		// double check these
		Added:   []string{},
		Deleted: []string{},
	},
}

func Createv1UpgradeHandler(
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Add additional upgrade logic when needed

		/*
		 * migrations are run in module name alphabetical
		 * ascending order, except x/auth which is run last
		 */
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
