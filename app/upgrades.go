package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/sedaprotocol/seda-chain/app/upgrades"
	v018 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v0.1.8"
	v019 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v0.1.9"
	v1 "github.com/sedaprotocol/seda-chain/app/upgrades/testnet/v1"
)

// Upgrades list of chain upgrades
var Upgrades = []upgrades.Upgrade{
	// testnet
	v1.Upgrade,
	// mainnet
	v018.Upgrade,
	v019.Upgrade,
}

func (app *App) setupUpgrades() {
	app.setUpgradeHandlers()
	app.setUpgradeStoreLoaders()
}

func (app *App) setUpgradeHandlers() {
	keepers := app.AppKeepers

	// register all upgrade handlers
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

	// Add new modules here when needed

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// register store loader for current upgrade
	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
			break
		}
	}
}
