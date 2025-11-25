package testutil

import (
	"encoding/json"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/x/core/types"
)

func (dr *TestDR) ExecuteDataRequestFlowContract(f *Fixture, numCommits, numReveals int, timeout bool) string {
	dr.PostDataRequestContract(f)
	dr.CommitDataRequestContract(f, numCommits, nil)
	dr.RevealDataRequestContract(f, numReveals, nil)

	if timeout {
		timeoutBlocks := defaultCommitTimeoutBlocks
		if numCommits == dr.ReplicationFactor {
			timeoutBlocks = defaultRevealTimeoutBlocks
		}

		for range timeoutBlocks {
			f.AddBlock()
		}
	}
	return dr.postedID
}

func (dr *TestDR) PostDataRequestContract(f *Fixture) {
	execGasLimit := uint64(160e15)
	tallyGasLimit := uint64(300e12)
	// (exec gas limit + tally gas limit) * gas price
	attachedFunds := math.NewIntFromUint64(execGasLimit).Add(math.NewIntFromUint64(tallyGasLimit)).Mul(math.NewInt(2000))

	resJSON := f.executeCoreContract(
		f.Creator.Address(),
		testutil.PostDataRequestMsgContract(dr.ExecProgHash, dr.TallyProgHash, dr.Memo, dr.ReplicationFactor, execGasLimit, tallyGasLimit),
		sdk.NewCoins(sdk.NewCoin(BondDenom, attachedFunds)),
	)

	type PostDataRequestResponse struct {
		DrID   string `json:"dr_id"`
		Height uint64 `json:"height"`
	}
	var res PostDataRequestResponse
	err := json.Unmarshal(resJSON, &res)
	require.NoError(f.tb, err)
	dr.postedID = res.DrID
	dr.postedHeight = res.Height
}

func (dr *TestDR) PostDataRequestContractShouldErr(f *Fixture, errMsg string) {
	execGasLimit := uint64(160e15)
	tallyGasLimit := uint64(300e12)
	// (exec gas limit + tally gas limit) * gas price
	attachedFunds := math.NewIntFromUint64(execGasLimit).Add(math.NewIntFromUint64(tallyGasLimit)).Mul(math.NewInt(2000))

	f.executeCoreContractShouldErr(
		f.Creator.Address(),
		testutil.PostDataRequestMsgContract(dr.ExecProgHash, dr.TallyProgHash, dr.Memo, dr.ReplicationFactor, execGasLimit, tallyGasLimit),
		sdk.NewCoins(sdk.NewCoin(BondDenom, attachedFunds)),
		errMsg,
	)
}

// CommitDataRequestContract executes commits from the first n stakers using the Core
// Contract messages. Corresponding reveal messages are stored in the TestDR object.
// If stakerIndices is not nil, stakers at the specified indices will be used.
func (dr *TestDR) CommitDataRequestContract(f *Fixture, n int, stakerIndices []int) {
	require.LessOrEqual(f.tb, n, len(f.Stakers))
	if stakerIndices != nil {
		require.Equal(f.tb, n, len(stakerIndices))
	}

	gasLimit := dr.CommitGasLimit
	if dr.CommitGasLimit == 0 {
		// TODO Use gastimation formula
		gasLimit = 500000
	}

	drID := dr.postedID
	drHeight := dr.postedHeight

	revealBody := types.RevealBody{
		DrID:         drID,
		Reveal:       dr.Reveal,
		GasUsed:      dr.GasUsed,
		ExitCode:     uint32(dr.ExitCode),
		ProxyPubKeys: dr.ProxyPubKeys,
	}

	var indices []int
	if stakerIndices != nil {
		indices = stakerIndices
	} else {
		indices = make([]int, n)
		for i := 0; i < n; i++ {
			indices[i] = i
		}
	}

	for _, i := range indices {
		revealMsg, commitment, _ := f.createRevealMsgContract(f.Stakers[i], revealBody)
		proof := f.generateCommitProof(f.Stakers[i].PrivKey, drID, commitment, drHeight)
		commitMsg := testutil.CommitMsgContract(drID, commitment, f.Stakers[i].PubKey, proof)

		f.executeCommitOrRevealContract(f.Stakers[i].Address, commitMsg, gasLimit)
		dr.revealMsgsContract = append(dr.revealMsgsContract, revealMsg)
	}
}

// RevealDataRequestContract executes reveal messages stored from execution of
// CommitDataRequestContract using the first n stakers.
// If stakerIndices is not nil, stakers at the specified indices will be used.
// After execution, stored reveal messages are cleared.
func (dr *TestDR) RevealDataRequestContract(f *Fixture, n int, stakerIndices []int) {
	require.LessOrEqual(f.tb, n, len(f.Stakers))
	require.LessOrEqual(f.tb, n, len(dr.revealMsgsContract))
	if stakerIndices != nil {
		require.Equal(f.tb, n, len(stakerIndices))
	}

	gasLimit := dr.RevealGasLimit
	if dr.RevealGasLimit == 0 {
		// TODO Use gastimation formula
		gasLimit = 500000
	}

	// Determine which staker indices to use.
	var indices []int
	if stakerIndices != nil {
		indices = stakerIndices
	} else {
		indices = make([]int, n)
		for i := 0; i < n; i++ {
			indices[i] = i
		}
	}

	for _, i := range indices {
		f.executeCommitOrRevealContract(f.Stakers[i].Address, dr.revealMsgsContract[i], gasLimit)
	}
	dr.revealMsgsContract = nil
}
