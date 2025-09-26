package testutil

import (
	"encoding/json"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

type TestDR struct {
	ExecProgHash      []byte
	TallyProgHash     []byte
	Memo              string
	ReplicationFactor int

	// To simulate reveal contents reported by executors:
	Reveal       []byte
	ProxyPubKeys []string
	GasUsed      uint64
	ExitCode     byte

	// Gas limits for commit & reveal txs
	CommitGasLimit uint64 // defaults to 100000 if not set or set to 0
	RevealGasLimit uint64 // defaults to 100000 if not set or set to 0

	// For tracking intermediate results
	postedID     string   // set after post (DR ID)
	postedHeight uint64   // set after post (height)
	revealMsgs   [][]byte // set after commit to be used in reveal
}

// NewTestDR instantiates a new TestDR object with given parameters.
func NewTestDR(execProgHash, tallyProgHash, reveal []byte, memo string, gasUsed uint64, exitCode byte, proxyPubKeys []string, rf int) TestDR {
	return TestDR{
		ExecProgHash:      execProgHash,
		TallyProgHash:     tallyProgHash,
		Memo:              memo,
		ReplicationFactor: rf,
		Reveal:            reveal,
		ProxyPubKeys:      proxyPubKeys,
		GasUsed:           gasUsed,
		ExitCode:          exitCode,
	}
}

// NewTestDRWithRandomPrograms instantiates a new TestDR object with wasm
// programs randomly selected from the testwasms catalog.
func NewTestDRWithRandomPrograms(reveal []byte, memo string, gasUsed uint64, exitCode byte, proxyPubKeys []string, rf int, blockTime time.Time) TestDR {
	randIndex := rand.Intn(len(testwasms.TestWasms))
	execProgHash := wasmstoragetypes.NewOracleProgram(testwasms.TestWasms[randIndex], blockTime).Hash
	randIndex = rand.Intn(len(testwasms.TestWasms))
	tallyProgHash := wasmstoragetypes.NewOracleProgram(testwasms.TestWasms[randIndex], blockTime).Hash

	return TestDR{
		ExecProgHash:      execProgHash,
		TallyProgHash:     tallyProgHash,
		Memo:              memo,
		ReplicationFactor: rf,
		Reveal:            reveal,
		ProxyPubKeys:      proxyPubKeys,
		GasUsed:           gasUsed,
		ExitCode:          exitCode,
	}
}

func (dr *TestDR) SetCommitRevealGasLimits(commitGasLimit, revealGasLimit uint64) {
	dr.CommitGasLimit = commitGasLimit
	dr.RevealGasLimit = revealGasLimit
}

func (dr *TestDR) PostDataRequest(f *Fixture) {
	amount, ok := math.NewIntFromString("200600000000000000000")
	require.True(f.tb, ok)

	resJSON := f.executeCoreContract(
		f.Creator.Address(),
		testutil.PostDataRequestMsg(dr.ExecProgHash, dr.TallyProgHash, dr.Memo, dr.ReplicationFactor),
		sdk.NewCoins(sdk.NewCoin(BondDenom, amount)),
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

// commitDataRequest executes a commit for each of the given stakers and
// returns a list of corresponding reveal messages.
func (dr *TestDR) CommitDataRequest(f *Fixture, numCommits int) {
	require.LessOrEqual(f.tb, numCommits, len(f.Stakers))

	gasLimit := dr.CommitGasLimit
	if dr.CommitGasLimit == 0 {
		// TODO Use gastimation formula
		gasLimit = 200000
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

	var revealMsgs [][]byte
	for i := 0; i < numCommits; i++ {
		revealMsg, commitment, _ := f.createRevealMsg(f.Stakers[i], revealBody)

		proof := f.generateCommitProof(f.Stakers[i].Key, drID, commitment, drHeight)

		commitMsg := testutil.CommitMsg(drID, commitment, f.Stakers[i].PubKey, proof)
		f.executeCommitOrReveal(f.Stakers[i].Address, commitMsg, gasLimit)

		revealMsgs = append(revealMsgs, revealMsg)
	}

	dr.revealMsgs = revealMsgs
}

// executeReveals executes a list of reveal messages, one by one using
// the staker addresses.
func (dr *TestDR) ExecuteReveals(f *Fixture, numReveals int) {
	require.LessOrEqual(f.tb, numReveals, len(dr.revealMsgs))
	require.LessOrEqual(f.tb, numReveals, len(f.Stakers))

	gasLimit := dr.RevealGasLimit
	if dr.RevealGasLimit == 0 {
		// TODO Use gastimation formula
		gasLimit = 200000
	}

	for i := 0; i < numReveals; i++ {
		f.executeCommitOrReveal(f.Stakers[i].Address, dr.revealMsgs[i], gasLimit)
	}
}

func (dr *TestDR) ExecuteDataRequestFlow(f *Fixture, numCommits, numReveals int, timeout bool) string {
	dr.PostDataRequest(f)
	dr.CommitDataRequest(f, numCommits)
	dr.ExecuteReveals(f, numReveals)

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

func (dr TestDR) GetDataRequestID() string {
	return dr.postedID
}
