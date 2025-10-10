package datarequesttests

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestBasicPayoutWorks(t *testing.T) {
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
	executor1PayoutAmount, _ := executor1.CalculatePayoutAmount(5, dr.GasPrice, false)

	posterBalance := poster.Balance()
	// How much the tally vm uses to run the hello world program
	helloWorldTallyGas := math.NewIntFromUint64(5000003670000)
	expectedBurnAmount := math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Add(helloWorldTallyGas).Mul(dr.GasPrice)
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount)
	require.Equal(t, expectedBalance.String(), posterBalance.String())
}

func TestReducedPayoutWorks(t *testing.T) {
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
	executor1PayoutAmount, executor1Burned := executor1.CalculatePayoutAmount(5, dr.GasPrice, true)
	reducedPayoutBurn = reducedPayoutBurn.Add(executor1Burned)

	posterBalance := poster.Balance()
	expectedBurnAmount := (math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Mul(dr.GasPrice)).Add((reducedPayoutBurn.Mul(dr.GasPrice)))
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount)
	require.Equal(t, expectedBalance.String(), posterBalance.String())
}

func TestReplicationFactorSameGasUsedPayoutWorks(t *testing.T) {
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
	executor1PayoutAmount, _ := executor1.CalculatePayoutAmount(5, dr.GasPrice, false)
	executor2PayoutAmount, _ := executor2.CalculatePayoutAmount(5, dr.GasPrice, false)
	executor3PayoutAmount, _ := executor3.CalculatePayoutAmount(5, dr.GasPrice, false)
	t.Log("payoutAmounts", executor1PayoutAmount, executor2PayoutAmount, executor3PayoutAmount)

	posterBalance := poster.Balance()
	// How much the tally vm uses to run the hello world program
	helloWorldTallyGas := math.NewIntFromUint64(5000003670000)
	expectedBurnAmount := math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Add(helloWorldTallyGas).Mul(dr.GasPrice)
	t.Log("tallygasused", math.NewIntFromUint64(tallyConfig.FilterGasCostNone+tallyConfig.BaseGasCost).Add(helloWorldTallyGas).String())
	// TODO: our tally gas used is wrong for this test... why?
	// actually used: 6000009310000
	// expected use: 6000003770000
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount).Sub(executor2PayoutAmount).Sub(executor3PayoutAmount)
	require.Equal(t, expectedBalance.String(), posterBalance.String())

}

func TestReplicationFactorDifferentGasUsedPayoutWorks(t *testing.T) {
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

	// executor 2 and 3 commits and with 2 gas used
	reveal2 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       2,
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

	drr, err := f.BatchingKeeper.GetDataResult(f.Context(), postDrResult.DrID, uint64(postDrResult.Height))
	t.Log(drr)
	require.NoError(t, err)

	// payout is reduced account for that.
	tallyConfig, err := f.CoreKeeper.GetTallyConfig(f.Context())
	require.NoError(t, err)

	// We don't expect any reduced payout in this test
	// TODO:... math is different from single executor test above need to account for median gas used, and highest and lowest?
	executor1PayoutAmount, _ := executor1.CalculatePayoutAmount(5, dr.GasPrice, false)
	executor2PayoutAmount, _ := executor2.CalculatePayoutAmount(2, dr.GasPrice, false)
	executor3PayoutAmount, _ := executor3.CalculatePayoutAmount(2, dr.GasPrice, false)

	posterBalance := poster.Balance()
	// How much the tally vm uses to run the hello world program
	helloWorldTallyGas := math.NewIntFromUint64(5000003670000)
	expectedBurnAmount := math.NewIntFromUint64(tallyConfig.FilterGasCostNone + tallyConfig.BaseGasCost).Add(helloWorldTallyGas).Mul(dr.GasPrice)
	expectedBalance := f.SedaToAseda(22).Sub(expectedBurnAmount).Sub(executor1PayoutAmount).Sub(executor2PayoutAmount).Sub(executor3PayoutAmount)
	require.Equal(t, expectedBalance.String(), posterBalance.String())

}
