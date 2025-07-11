package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/sedaprotocol/seda-chain/app/upgrades"
	v017 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v0.1.7"
	v018 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v0.1.8"
	v019 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v0.1.9"
	v1 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v1"
	v1rc4 "github.com/sedaprotocol/seda-chain/app/upgrades/testnet/v1.0.0-rc.4"
	v1rc6 "github.com/sedaprotocol/seda-chain/app/upgrades/testnet/v1.0.0-rc.6"
)

// Upgrades is a list of currently supported upgrades.
var Upgrades = []upgrades.Upgrade{
	v1rc4.Upgrade,
	v1rc6.Upgrade,
	v1.Upgrade,
	v017.Upgrade,
	v018.Upgrade,
	v019.Upgrade,
}

func (app *App) setupUpgrades() {
	app.setUpgradeHandlers()
	app.setUpgradeStoreLoaders()
}

func (app *App) setUpgradeHandlers() {
	keepers := app.AppKeepers

	// Register upgrade handlers.
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				&keepers,
			),
		)
	}
}

func (app *App) setUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	// Add new modules here when needed.

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// Register store loader for current upgrade.
	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
			break
		}
	}
}
