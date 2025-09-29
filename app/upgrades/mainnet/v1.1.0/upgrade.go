package upgrade

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sedaprotocol/seda-chain/app/keepers"
	"github.com/sedaprotocol/seda-chain/app/upgrades"
	v1 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v1"
	coretypes "github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/sedaprotocol/seda-chain/x/wasm"
)

const (
	UpgradeName = "v1.1.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added:   []string{coretypes.StoreKey},
		Deleted: []string{v1.TallyStoreKey},
	},
}

func CreateUpgradeHandler(
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(context context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(context)

		// Migrate Core Contract state to x/core.
		contractAddr, err := keepers.WasmStorageKeeper.GetCoreContractAddr(ctx)
		if err != nil {
			return nil, err
		}

		var owner string
		var allowlist []string
		var stakers []coretypes.Staker
		stakerToSeq := make(map[string]uint64)
		keepers.WasmKeeper.IterateContractState(ctx, contractAddr, func(key []byte, value []byte) bool {
			keyStr := string(key)
			switch {
			case strings.Contains(keyStr, "owner"):
				err := json.Unmarshal(value, &owner)
				if err != nil {
					panic(fmt.Sprintf("failed to unmarshal owner: %s", owner))
				}

			case strings.Contains(keyStr, "allowlist"):
				pubkey := strings.Split(keyStr, "allowlist")[1]
				allowlist = append(allowlist, hex.EncodeToString([]byte(pubkey)))

			case strings.Contains(keyStr, "data_request_executors_stakers"):
				pubkey := strings.Split(keyStr, "data_request_executors_stakers")[1]

				var staker wasm.StakerResponse
				err := json.Unmarshal(value, &staker)
				if err != nil {
					panic(fmt.Sprintf("failed to unmarshal staker: %s", staker))
				}

				tokensStaked, ok := math.NewIntFromString(staker.TokensStaked)
				if !ok {
					panic(fmt.Sprintf("failed to parse staker tokens staked: %s, staker: %s", staker.TokensStaked, staker.PublicKey))
				}
				tokensPendingWithdrawal, ok := math.NewIntFromString(staker.TokensPendingWithdrawal)
				if !ok {
					panic(fmt.Sprintf("failed to parse pending withdrawal: %s, staker: %s", staker.TokensStaked, staker.PublicKey))
				}

				stakers = append(stakers, coretypes.Staker{
					PublicKey:         hex.EncodeToString([]byte(pubkey)),
					Memo:              string(staker.Memo),
					Staked:            tokensStaked,
					PendingWithdrawal: tokensPendingWithdrawal,
				})

			case strings.Contains(keyStr, "account_seq"):
				pubkey := strings.Split(keyStr, "account_seq")[1]
				var seq uint64
				err := json.Unmarshal(value, &seq)
				if err != nil {
					panic(fmt.Sprintf("failed to unmarshal staker seq: %d, staker: %s", seq, pubkey))
				}
				stakerToSeq[hex.EncodeToString([]byte(pubkey))] = seq

			// Below keys should not be present in the Core Contract state
			// because the data request pool should have been drained.
			case strings.Contains(keyStr, "dr_staked_funds"):
				panic(fmt.Sprintf("unexpected key: %s", keyStr))

			case strings.Contains(keyStr, "data_request_pool"):
				if !strings.Contains(keyStr, "data_request_pool_committing_len") &&
					!strings.Contains(keyStr, "data_request_pool_revealing_len") &&
					!strings.Contains(keyStr, "data_request_pool_tallying_len") {
					panic(fmt.Sprintf("unexpected key: %s", keyStr))
				}
			}

			return false
		})

		err = keepers.CoreKeeper.SetOwner(ctx, owner)
		if err != nil {
			return nil, err
		}

		for _, pubkey := range allowlist {
			err = keepers.CoreKeeper.AddToAllowlist(ctx, pubkey)
			if err != nil {
				return nil, err
			}
		}

		for _, staker := range stakers {
			staker.SequenceNum = stakerToSeq[staker.PublicKey]
			err = keepers.CoreKeeper.SetStaker(ctx, staker)
			if err != nil {
				return nil, err
			}
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
