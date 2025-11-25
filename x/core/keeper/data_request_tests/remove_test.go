package datarequesttests

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestBasicPayout(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create the dr poster
	poster := f.CreateTestAccount("poster", 22)
	// create the dr executors
	executor1 := f.CreateStakedTestAccount("executor1", 22, 1)

	dr := poster.CalculateDrIDAndArgs("1", 1)
	dr.ExecProgramID = f.DeployedOPs["hello_world"]
	dr.TallyProgramID = f.DeployedOPs["hello_world"]
	postDrResult, err := poster.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// executor 1 commits with 5 gas used
	reveal1 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       5,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	executor1RevealMsg := executor1.CreateRevealMsg(reveal1)
	_, err = executor1.CommitResult(executor1RevealMsg)
	require.NoError(t, err)

	// reveal phase
	_, err = executor1.RevealResult(executor1RevealMsg)
	require.NoError(t, err)

	// Move to next block so the dr can be removed
	f.AdvanceBlocks(1)

	_, err = f.BatchingKeeper.GetDataResult(f.Context(), postDrResult.DrID, uint64(postDrResult.Height))
	require.NoError(t, err)

	// payout is reduced account for that.
	tallyConfig, err := f.CoreKeeper.GetTallyConfig(f.Context())
	require.NoError(t, err)

	// We don't expect any reduced payout in this test
	executor1PayoutAmount, _ := executor1.CheckGasPayoutUniformCase(5, dr.GasPrice, false)

	// How much the tally vm uses to run the hello world program
	expectedTallyVMGas := math.NewIntFromUint64(5000003670000) // fixed value

	expectedBurnAmount := math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Add(expectedTallyVMGas).Mul(dr.GasPrice)
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount)
	require.Equal(t, expectedBalance.String(), poster.Balance().String())
}

func TestReducedPayout(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create the dr poster
	poster := f.CreateTestAccount("poster", 22)
	// create the dr executors
	executor1 := f.CreateStakedTestAccount("executor1", 22, 1)

	dr := poster.CalculateDrIDAndArgs("1", 1)
	postDrResult, err := poster.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// executor 1 commits with 5 gas used
	reveal1 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       5,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	executor1RevealMsg := executor1.CreateRevealMsg(reveal1)
	_, err = executor1.CommitResult(executor1RevealMsg)
	require.NoError(t, err)

	// reveal phase
	_, err = executor1.RevealResult(executor1RevealMsg)
	require.NoError(t, err)

	// Move to next block so the dr can be removed
	f.AdvanceBlocks(1)

	_, err = f.BatchingKeeper.GetDataResult(f.Context(), postDrResult.DrID, uint64(postDrResult.Height))
	require.NoError(t, err)

	// payout is reduced account for that.
	tallyConfig, err := f.CoreKeeper.GetTallyConfig(f.Context())
	require.NoError(t, err)

	reducedPayoutBurn := math.ZeroInt()
	executor1PayoutAmount, executor1Burned := executor1.CheckGasPayoutUniformCase(5, dr.GasPrice, true)
	reducedPayoutBurn = reducedPayoutBurn.Add(executor1Burned)

	posterBalance := poster.Balance()
	expectedBurnAmount := (math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Mul(dr.GasPrice)).Add((reducedPayoutBurn.Mul(dr.GasPrice)))
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount)
	require.Equal(t, expectedBalance.String(), posterBalance.String())
}

func TestUniformGasUsedPayout(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create the dr poster
	poster := f.CreateTestAccount("poster", 22)
	// create the dr executors
	executor1 := f.CreateStakedTestAccount("executor1", 22, 1)
	executor2 := f.CreateStakedTestAccount("executor2", 22, 1)
	executor3 := f.CreateStakedTestAccount("executor3", 22, 1)

	dr := poster.CalculateDrIDAndArgs("1", 3)
	dr.ExecProgramID = f.DeployedOPs["hello_world"]
	dr.TallyProgramID = f.DeployedOPs["hello_world"]
	postDrResult, err := poster.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	reveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       5,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	executor1RevealMsg := executor1.CreateRevealMsg(reveal)
	_, err = executor1.CommitResult(executor1RevealMsg)
	require.NoError(t, err)

	executor2RevealMsg := executor2.CreateRevealMsg(reveal)
	_, err = executor2.CommitResult(executor2RevealMsg)
	require.NoError(t, err)
	executor3RevealMsg := executor3.CreateRevealMsg(reveal)
	_, err = executor3.CommitResult(executor3RevealMsg)
	require.NoError(t, err)

	// reveal phase
	_, err = executor1.RevealResult(executor1RevealMsg)
	require.NoError(t, err)
	_, err = executor2.RevealResult(executor2RevealMsg)
	require.NoError(t, err)
	_, err = executor3.RevealResult(executor3RevealMsg)
	require.NoError(t, err)

	// Move to next block so the dr can be removed
	f.AdvanceBlocks(1)

	_, err = f.BatchingKeeper.GetDataResult(f.Context(), postDrResult.DrID, uint64(postDrResult.Height))
	require.NoError(t, err)

	// payout is reduced account for that.
	tallyConfig, err := f.CoreKeeper.GetTallyConfig(f.Context())
	require.NoError(t, err)

	// We don't expect any reduced payout in this test
	executor1PayoutAmount, _ := executor1.CheckGasPayoutUniformCase(5, dr.GasPrice, false)
	executor2PayoutAmount, _ := executor2.CheckGasPayoutUniformCase(5, dr.GasPrice, false)
	executor3PayoutAmount, _ := executor3.CheckGasPayoutUniformCase(5, dr.GasPrice, false)

	expectedTallyVMGas := math.NewIntFromUint64(5000009210000) // fixed value

	expectedBurnAmount := math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Add(expectedTallyVMGas).Mul(dr.GasPrice)
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount).Sub(executor2PayoutAmount).Sub(executor3PayoutAmount)
	require.Equal(t, expectedBalance.String(), poster.Balance().String())
}

func TestReplicationFactorDifferentGasUsedPayout(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create the dr poster
	poster := f.CreateTestAccount("poster", 22)
	// create the dr executors
	executor1 := f.CreateStakedTestAccount("executor1", 22, 1)
	executor2 := f.CreateStakedTestAccount("executor2", 22, 1)
	executor3 := f.CreateStakedTestAccount("executor3", 22, 1)

	dr := poster.CalculateDrIDAndArgs("1", 3)
	dr.ExecProgramID = f.DeployedOPs["hello_world"]
	dr.TallyProgramID = f.DeployedOPs["hello_world"]
	postDrResult, err := poster.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// executor 1 commits with 500 gas used
	reveal1 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       500,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	executor1RevealMsg := executor1.CreateRevealMsg(reveal1)
	_, err = executor1.CommitResult(executor1RevealMsg)
	require.NoError(t, err)

	// executor 2 and 3 commits and with 200 gas used
	reveal2 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       200,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	executor2RevealMsg := executor2.CreateRevealMsg(reveal2)
	_, err = executor2.CommitResult(executor2RevealMsg)
	require.NoError(t, err)
	executor3RevealMsg := executor3.CreateRevealMsg(reveal2)
	_, err = executor3.CommitResult(executor3RevealMsg)
	require.NoError(t, err)

	// reveal phase
	_, err = executor1.RevealResult(executor1RevealMsg)
	require.NoError(t, err)
	_, err = executor2.RevealResult(executor2RevealMsg)
	require.NoError(t, err)
	_, err = executor3.RevealResult(executor3RevealMsg)
	require.NoError(t, err)

	// Move to next block so the dr can be removed
	f.AdvanceBlocks(1)

	tallyConfig, err := f.CoreKeeper.GetTallyConfig(f.Context())
	require.NoError(t, err)

	// Canonical gas calculation:
	// median = 200
	// totalShares = median*(RF - 1) + low*2 = 2*200 + 2*200 = 800
	// totalUsedGas = median*(RF - 1) + min(low*2, median) = 2*200 + min(2*200, 200) = 600
	// lowestGasUsed = (low*2)*totalUsedGas/totalShares = 2*200*600/800 = 300
	// regGasUsed = median*totalUsedGas/totalShares = 200*600/800 = 150
	expPayExec1 := math.NewInt(300).Mul(dr.GasPrice)
	expPayExec2 := math.NewInt(150).Mul(dr.GasPrice)
	expPayExec3 := math.NewInt(150).Mul(dr.GasPrice)

	expectedTallyVMGas := math.NewIntFromUint64(5000009270000) // fixed value

	expectedBurnAmount := math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Add(expectedTallyVMGas).Mul(dr.GasPrice)
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(expPayExec1).Sub(expPayExec2).Sub(expPayExec3)
	require.Equal(t, expectedBalance.String(), poster.Balance().String())
}
