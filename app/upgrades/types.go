package upgrades

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sedaprotocol/seda-chain/app/keepers"
)

type ModuleManager interface {
	RunMigrations(ctx context.Context, cfg module.Configurator, fromVM module.VersionMap) (module.VersionMap, error)
	GetVersionMap() module.VersionMap
}

type Upgrade struct {
	UpgradeName          string
	CreateUpgradeHandler func(ModuleManager, module.Configurator, *keepers.AppKeepers) upgradetypes.UpgradeHandler
	StoreUpgrades        storetypes.StoreUpgrades
}

// NewDefaultUpgrade creates a new Upgrade object under the given name with a
// simple upgrade handler that calls ModuleManager.RunMigrations without any
// additional logic.
func NewDefaultUpgrade(upgradeName string) Upgrade {
	return Upgrade{
		UpgradeName:          upgradeName,
		CreateUpgradeHandler: DefaultUpgradeHandler,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added:   []string{},
			Deleted: []string{},
		},
	}
}

func DefaultUpgradeHandler(
	mm ModuleManager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		/*
		 * migrations are run in module name alphabetical
		 * ascending order, except x/auth which is run last
		 */
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
