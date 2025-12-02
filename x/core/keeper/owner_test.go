package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestGetOwner(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	someone := f.CreateTestAccount("someone", 10_000)

	resp, err := someone.GetOwner()
	require.NoError(t, err)
	require.Equal(t, f.Creator.Address(), resp.Owner)
}

func TestGetNoPendingOwner(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	someone := f.CreateTestAccount("someone", 10_000)

	resp, err := someone.GetPendingOwner()
	require.Empty(t, resp.PendingOwner)
	require.NoError(t, err)
}

func TestNonOwnerCannotTransferOwnership(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	someone := f.CreateTestAccount("someone", 10_000)

	_, err := someone.TransferOwnership("seda1zcds6ws7l0e005h3xrmg5tx0378nyg8gtmn64f")
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}

func TestTwoStepOwnershipTransfer(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	newOwner := f.CreateTestAccount("newowner", 10_000)

	// Owner transfers ownership to newOwner
	_, err := f.Creator.TransferOwnership(newOwner.Address())
	require.NoError(t, err)

	// Check pending owner is newOwner
	resp, err := newOwner.GetPendingOwner()
	require.NoError(t, err)
	require.Equal(t, newOwner.Address(), resp.PendingOwner)

	// new Owner accepts ownership
	_, err = newOwner.AcceptOwnership()
	require.NoError(t, err)

	// Check owner is now newOwner
	ownerResp, err := newOwner.GetOwner()
	require.NoError(t, err)
	require.Equal(t, newOwner.Address(), ownerResp.Owner)

	// Check no pending owner
	resp, err = f.Creator.GetPendingOwner()
	require.Empty(t, resp.PendingOwner)
	require.NoError(t, err)
}

func TestNonPendingOwnerCannotAcceptOwnership(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	newOwner := f.CreateTestAccount("newowner", 10_000)
	someone := f.CreateTestAccount("someone", 10_000)

	// Owner transfers ownership to newOwner
	_, err := f.Creator.TransferOwnership(newOwner.Address())
	require.NoError(t, err)

	// someone tries to accept ownership
	_, err = someone.AcceptOwnership()
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// Check owner is still original owner
	ownerResp, err := f.Creator.GetOwner()
	require.NoError(t, err)
	require.Equal(t, f.Creator.Address(), ownerResp.Owner)

	// Check pending owner is still newOwner
	pendingResp, err := newOwner.GetPendingOwner()
	require.NoError(t, err)
	require.Equal(t, newOwner.Address(), pendingResp.PendingOwner)
}

func TestOwnerCanChangeParams(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// Owner can change params
	stakingConfig := types.DefaultParams().StakingConfig
	stakingConfig.MinimumStake = math.NewInt(1)

	_, err := f.Creator.SetStakingConfig(*stakingConfig)
	require.NoError(t, err)

	// Check params were updated
	paramsResp, err := f.Creator.GetStakingConfig()
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), paramsResp.StakingConfig.MinimumStake)
}

func TestNonOwnerCannotChangeParams(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	someone := f.CreateTestAccount("someone", 10_000)

	stakingConfig := types.DefaultParams().StakingConfig
	stakingConfig.MinimumStake = math.NewInt(1)

	_, err := someone.SetStakingConfig(*stakingConfig)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// Check params were not updated
	paramsResp, err := someone.GetStakingConfig()
	require.NoError(t, err)
	require.NotEqual(t, math.NewInt(1), paramsResp.StakingConfig.MinimumStake)
}

func TestNewOwnerCanChangeParams(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	newOwner := f.CreateTestAccount("newowner", 10_000)

	// Owner transfers ownership to newOwner
	_, err := f.Creator.TransferOwnership(newOwner.Address())
	require.NoError(t, err)

	// new Owner accepts ownership
	_, err = newOwner.AcceptOwnership()
	require.NoError(t, err)

	// new Owner can change params
	stakingConfig := types.DefaultParams().StakingConfig
	stakingConfig.MinimumStake = math.NewInt(1)

	_, err = newOwner.SetStakingConfig(*stakingConfig)
	require.NoError(t, err)

	// Check params were updated
	paramsResp, err := newOwner.GetStakingConfig()
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), paramsResp.StakingConfig.MinimumStake)
}

func TestPauseBasics(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// initially not paused
	paused, err := f.CoreKeeper.IsPaused(f.Context())
	require.NoError(t, err)
	require.False(t, paused)

	// pause the contract
	_, err = f.Creator.Pause()
	require.NoError(t, err)

	// cannot pause again
	_, err = f.Creator.Pause()
	require.Error(t, err)

	// unpause the contract
	_, err = f.Creator.Unpause()
	require.NoError(t, err)

	// cannot unpause again
	_, err = f.Creator.Unpause()
	require.Error(t, err)
}

func TestOnlyOwnerCanPause(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateTestAccount("alice", 10_000)

	// non-owner cannot pause
	_, err := alice.Pause()
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)

	// non-owner cannot unpause
	_, err = alice.Unpause()
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}

func TestCanStillUpdateParamsIfPaused(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// pause the contract
	_, err := f.Creator.Pause()
	require.NoError(t, err)

	// Owner can change params
	stakingConfig := types.DefaultParams().StakingConfig
	stakingConfig.MinimumStake = math.NewInt(1)

	_, err = f.Creator.SetStakingConfig(*stakingConfig)
	require.NoError(t, err)

	// Check params were updated
	paramsResp, err := f.Creator.GetStakingConfig()
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), paramsResp.StakingConfig.MinimumStake)
}
