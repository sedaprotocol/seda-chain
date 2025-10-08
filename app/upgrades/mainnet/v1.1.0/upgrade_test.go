package upgrade_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/exp/rand"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sedaprotocol/seda-chain/app/keepers"
	v1_1_0 "github.com/sedaprotocol/seda-chain/app/upgrades/mainnet/v1.1.0"
	"github.com/sedaprotocol/seda-chain/app/upgrades/testutil"
	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	coretest "github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/wasm"
)

// TestCoreContractUpgrade populates the Core Contract with data requests,
// drains the data request pool, and executes upgrade at height 100.
// After upgrade execution, it ensures that owner, allowlist, and stakers are
// migrated from the Core Contract.
func TestCoreContractUpgrade(t *testing.T) {
	// Setup with arbitrary data
	f := coretest.InitFixture(t, true, testwasms.CoreContractUpgradeReadyWasm())
	stakers := f.AddStakers(t, 15)

	totalDRs := 250
	drIDs := make([]string, totalDRs)
	rf := 3
	var postedDRs, committedDRs, revealedDRs int
	for j := 0; j < totalDRs; j++ {
		// Decide whether to post, post + commit, or post + commit + reveal.
		var numCommits, numReveals int
		switch j % 3 {
		case 0:
			postedDRs++
			numCommits = rand.Intn(rf)
		case 1:
			committedDRs++
			numCommits = rf
			numReveals = rand.Intn(rf)
		case 2:
			revealedDRs++
			numCommits = rf
			numReveals = rf
		}

		dr := coretest.NewRandomTestDR(f, rf)
		drIDs[j] = dr.GetDataRequestID()

		dr.PostDataRequest(f)
		dr.CommitDataRequest(f, numCommits, nil)
		dr.ExecuteReveals(f, numReveals, nil)
	}

	res := f.DrainDataRequestPool(100)
	require.Nil(t, res)

	// Post DR should fail once we reach block 40
	// upgrade_height - (commitTimeout + revealTimeout + backupDelay)
	// = 100 - (50 + 5 + 5) = 40
	dr := coretest.NewRandomTestDR(f, rf)
	dr.PostDataRequest(f)

	f.SimulateLegacyTallyEndblock(t, 39) // Block 39

	dr = coretest.NewRandomTestDR(f, rf)
	dr.PostDataRequest(f)

	f.SimulateLegacyTallyEndblock(t, 1) // Block 40

	dr = coretest.NewRandomTestDR(f, rf)
	dr.PostDataRequestShouldErr(f, "Cannot post request: Data request pool is draining")

	//
	// Upgrade handler execution
	//
	// Re-wire core contract so that the Core Contract is migrated.
	err := f.WasmStorageKeeper.CoreContractRegistry.Set(f.Context(), "")
	require.NoError(t, err)

	moduleManager := testutil.NewMockModuleManager(gomock.NewController(t))
	moduleManager.EXPECT().RunMigrations(gomock.Any(), gomock.Any(), gomock.Any()).Return(module.VersionMap{}, nil).Times(3)

	upgradeHandler := v1_1_0.CreateUpgradeHandler(
		moduleManager, nil,
		&keepers.AppKeepers{
			WasmStorageKeeper: f.WasmStorageKeeper,
			CoreKeeper:        f.CoreKeeper,
			WasmKeeper:        f.WasmViewKeeper.(*wasm.Keeper),
		},
	)

	// Upgrade handler should panic at block 60 due to non-empty data request pool.
	f.SimulateLegacyTallyEndblock(t, 20) // Block 60
	require.Panics(t, func() {
		upgradeHandler(f.Context(), upgradetypes.Plan{}, module.VersionMap{})
	})

	// Update handler executes successfully at block 100.
	f.SimulateLegacyTallyEndblock(t, 40) // Block 100
	_, err = upgradeHandler(f.Context(), upgradetypes.Plan{}, module.VersionMap{})
	require.NoError(t, err)

	//
	// Check the module state after the upgrade handler execution.
	//
	_, err = upgradeHandler(f.Context(), upgradetypes.Plan{}, module.VersionMap{})
	require.NoError(t, err)

	stakerCount, err := f.CoreKeeper.GetStakerCount(f.Context())
	require.NoError(t, err)
	require.Equal(t, uint32(15), stakerCount)

	for _, staker := range stakers {
		staker, err := f.CoreKeeper.GetStaker(f.Context(), staker.PubKey)
		require.NoError(t, err)
		require.Equal(t, staker.Staked, math.NewInt(1500000000000000000))

		isAllowlisted, err := f.CoreKeeper.IsAllowlisted(f.Context(), staker.PublicKey)
		require.NoError(t, err)
		require.True(t, isAllowlisted)
	}

	owner, err := f.CoreKeeper.GetOwner(f.Context())
	require.NoError(t, err)
	require.Equal(t, authtypes.NewModuleAddress(govtypes.ModuleName).String(), owner)
}
