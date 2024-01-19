package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/sedaprotocol/seda-chain/app/upgrades"
	v1 "github.com/sedaprotocol/seda-chain/app/upgrades/testnet/v1"
)

// Upgrades list of chain upgrades
var Upgrades = []upgrades.Upgrade{
	// testnet
	v1.Upgrade,
}

// RegisterUpgradeHandlers returns upgrade handlers
func (app *App) RegisterUpgradeHandlers() {
	keepers := upgrades.AppKeepers{
		AccountKeeper:         app.AccountKeeper,
		ConsensusParamsKeeper: app.ConsensusParamsKeeper,
		CapabilityKeeper:      app.CapabilityKeeper,
		IBCKeeper:             app.IBCKeeper,
		AuthzKeeper:           app.AuthzKeeper,
		BankKeeper:            app.BankKeeper,
		StakingKeeper:         app.StakingKeeper,
		SlashingKeeper:        app.SlashingKeeper,
		MintKeeper:            app.MintKeeper,
		DistrKeeper:           app.DistrKeeper,
		GovKeeper:             app.GovKeeper,
		CrisisKeeper:          app.CrisisKeeper,
		UpgradeKeeper:         app.UpgradeKeeper,
		EvidenceKeeper:        app.EvidenceKeeper,
		FeeGrantKeeper:        app.FeeGrantKeeper,
		GroupKeeper:           app.GroupKeeper,
		CircuitKeeper:         app.CircuitKeeper,
		ICAHostKeeper:         app.ICAHostKeeper,
		TransferKeeper:        app.TransferKeeper,
		WasmKeeper:            app.WasmKeeper,
		IBCFeeKeeper:          app.IBCFeeKeeper,
		ICAControllerKeeper:   app.ICAControllerKeeper,
	}

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
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
			break
		}
	}
}
