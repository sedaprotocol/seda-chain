package datarequesttests

import (
	"testing"

	"github.com/sedaprotocol/seda-chain/x/core/keeper/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	"github.com/stretchr/testify/require"
)

// TODO: These tests are hard to port will Pair on Monday
func TestBasicWorkflowWorks(t *testing.T) {
	f := testutil.InitFixture(t, false, nil)

	alice := f.CreateStakedTestAccount("alice", 22, 1)
	bob := f.CreateTestAccount("bob", 22)

	dr := bob.CalculateDrIdAndArgs("1", 1)
	postDrResult, err := bob.PostDataRequest(dr, 1, nil)
	require.NoError(t, err)

	// Alice commits and reveals result
	aliceReveal := &types.RevealBody{
		DrID:          postDrResult.DrID,
		DrBlockHeight: uint64(postDrResult.Height),
		Reveal:        testutil.RevealHelperFromString("10"),
		GasUsed:       0,
		ExitCode:      0,
		ProxyPubKeys:  []string{},
	}
	aliceRevealMsg := alice.CreateRevealMsg(aliceReveal)
	_, err = alice.CommitResult(aliceRevealMsg)
	require.NoError(t, err)
	_, err = alice.RevealResult(alice.CreateRevealMsg(aliceReveal))
	require.NoError(t, err)

	// Move to next block so the dr can be removed
	f.AdvanceBlocks(1)
}
