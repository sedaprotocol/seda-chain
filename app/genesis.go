package app

import (
	"crypto/rand"
	"encoding/json"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	vestingtypes "github.com/sedaprotocol/seda-chain/x/vesting/types"
)

// The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}

// RandomGenesisAccounts defines the default RandomGenesisAccountsFn used on the SDK.
// It creates a slice of BaseAccount, ContinuousVestingAccount and DelayedVestingAccount.
// NOTE: This function is a modified version of
// https://github.com/cosmos/cosmos-sdk/blob/7e6948f50cd4838a0161838a099f74e0b5b0213c/x/auth/simulation/genesis.go#L26
func RandomGenesisAccounts(simState *module.SimulationState) types.GenesisAccounts {
	genesisAccs := make(types.GenesisAccounts, len(simState.Accounts))
	var funderAddress string
	if int(simState.NumBonded) < len(simState.Accounts) {
		funderAddress = simState.Accounts[simState.NumBonded].Address.String() // acc at index NumBonded is designated funder
	}
	for i, acc := range simState.Accounts {
		bacc := types.NewBaseAccountWithAddress(acc.Address)

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) <= simState.NumBonded || simState.Rand.Intn(100) < 50 {
			genesisAccs[i] = bacc
			continue
		}

		maxInitialVesting := simState.InitialStake.Quo(math.NewInt(int64(len(simState.Accounts))))
		initialVestingAmount, err := rand.Int(rand.Reader, maxInitialVesting.BigInt())
		if err != nil {
			panic(err)
		}
		initialVesting := sdk.NewCoins(sdk.NewCoin(simState.BondDenom, math.NewIntFromBigInt(initialVestingAmount)))

		var endTime int64
		startTime := simState.GenTimestamp.Unix()
		// Allow for some vesting accounts to vest very quickly while others very slowly.
		if simState.Rand.Intn(100) < 50 {
			endTime = int64(simulation.RandIntBetween(simState.Rand, int(startTime)+1, int(startTime+(60*60*24*30))))
		} else {
			endTime = int64(simulation.RandIntBetween(simState.Rand, int(startTime)+1, int(startTime+(60*60*12))))
		}

		bva, err := sdkvestingtypes.NewBaseVestingAccount(bacc, initialVesting, endTime)
		if err != nil {
			panic(err)
		}

		genesisAccs[i] = vestingtypes.NewClawbackContinuousVestingAccountRaw(bva, startTime, funderAddress)
	}

	return genesisAccs
}
