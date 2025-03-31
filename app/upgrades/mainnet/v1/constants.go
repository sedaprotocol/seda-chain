package v1

import (
	"context"
	_ "embed"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

//go:embed core_contract_6827edee215c1fa01419cabeb7873db8eb7419ac.wasm
var wasmCode []byte

func CreateUpgradeHandler(
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)

		// Run module migrations.
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Additional upgrade logic for instantiating the Core Contract.
		securityGroupAddr := "seda1afk9zr2hn2jsac63h4hm60vl9z3e5u69gndzf7c99cqge3vzwjzs026662"
		adminAddr, err := sdk.AccAddressFromBech32(securityGroupAddr)
		if err != nil {
			return nil, err
		}

		codeID, _, err := keepers.WasmContractKeeper.Create(ctx, adminAddr, wasmCode, nil)
		if err != nil {
			return nil, err
		}
		contractAddr, _, err := keepers.WasmContractKeeper.Instantiate(
			ctx, codeID, adminAddr, adminAddr,
			[]byte(fmt.Sprintf(`{
				"token":"aseda",
				"owner": "%s", 
				"chain_id":"%s",
				"staking_config": {
					"minimum_stake": "10",
					"allowlist_enabled": true
				},
				"timeout_config": {
					"commit_timeout_in_blocks": 10,
					"reveal_timeout_in_blocks": 5
				}
			}`, securityGroupAddr, ctx.ChainID())),
			"label", nil)
		if err != nil {
			return nil, err
		}
		err = keepers.WasmStorageKeeper.CoreContractRegistry.Set(ctx, contractAddr.String())
		if err != nil {
			return nil, err
		}

		// Enable vote extension at current height.
		consensusParams, err := keepers.ConsensusParamsKeeper.ParamsStore.Get(ctx)
		if err != nil {
			return nil, err
		}
		consensusParams.Abci.VoteExtensionsEnableHeight = ctx.BlockHeight() + 1
		err = keepers.ConsensusParamsKeeper.ParamsStore.Set(ctx, consensusParams)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}
