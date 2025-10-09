package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

func TestProxyBasicPayoutWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	// create the dr poster
	poster := f.CreateTestAccount("poster", 22)
	// create the dr executors
	executor1 := f.CreateStakedTestAccount("executor1", 22, 1)
	executor2 := f.CreateStakedTestAccount("executor2", 22, 1)
	executor3 := f.CreateStakedTestAccount("executor3", 22, 1)

	// create the data proxy
	proxy := f.CreateProxyAccount("proxy")
	proxy.Register(2, nil)

	dr := poster.CalculateDrIDAndArgs("1", 3)
	postDrResult, err := poster.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// executor 1 commits with 5 gas used
	reveal1 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{proxy.PublicKeyHex()},
	}
	executor1RevealMsg := executor1.CreateRevealMsg(reveal1)
	_, err = executor1.CommitResult(executor1RevealMsg)
	require.NoError(t, err)

	// executor 2 and 3 commits and with 2 gas used
	reveal2 := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{proxy.PublicKeyHex()},
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

	// Check reward pending withdrawal for executors
	minimumDrPost := testutil.MinimumDrCost().Amount
	executor1Info, err := executor1.GetStaker()
	t.Log(executor1Info)
	require.NoError(t, err)
	require.Equal(t, minimumDrPost, executor1Info.Staker.PendingWithdrawal)

	executor2Info, err := executor2.GetStaker()
	require.NoError(t, err)
	require.Equal(t, minimumDrPost, executor2Info.Staker.PendingWithdrawal)

	executor3Info, err := executor3.GetStaker()
	require.NoError(t, err)
	require.Equal(t, minimumDrPost, executor3Info.Staker.PendingWithdrawal)

	// Check proxy was rewarded
	proxyBalance := proxy.Balance()
	require.Equal(t, f.SedaToAseda(100+2), proxyBalance.Int64())

	posterBalance := poster.Balance()
	// TODO: how much is burned???
	require.Equal(t, f.SedaToAseda(22-5-5-2-2-2), posterBalance.Int64())
}
