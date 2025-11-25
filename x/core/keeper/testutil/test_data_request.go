package testutil

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	postedID           string            // set after post (DR ID)
	postedHeight       uint64            // set after post (height)
	revealMsgsContract [][]byte          // set after commit to be used in reveal (Core Contract)
	revealMsgs         []types.MsgReveal // set after commit to be used in reveal (x/core module)
	revealProofs       []string          // set after commit to be used in reveal (x/core module)
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

func NewRandomTestDR(f *Fixture, rf int) TestDR {
	return NewTestDRWithRandomPrograms(
		[]byte("reveal"),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", rand.Int63()))),
		150000000000000000,
		0,
		[]string{},
		rf,
		f.Context().BlockTime(),
	)
}

func (dr *TestDR) ExecuteDataRequestFlow(f *Fixture, numCommits, numReveals int, timeout bool) string {
	dr.PostDataRequest(f)
	dr.CommitDataRequest(f, numCommits, nil)
	dr.RevealDataRequest(f, numReveals, nil)

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

func (dr *TestDR) SetCommitRevealGasLimits(commitGasLimit, revealGasLimit uint64) {
	dr.CommitGasLimit = commitGasLimit
	dr.RevealGasLimit = revealGasLimit
}

func (dr *TestDR) PostDataRequest(f *Fixture) {
	execGasLimit := uint64(160e15)
	tallyGasLimit := uint64(300e12)
	// (exec gas limit + tally gas limit) * gas price
	attachedFunds := math.NewIntFromUint64(execGasLimit).Add(math.NewIntFromUint64(tallyGasLimit)).Mul(math.NewInt(2000))

	res, err := f.CoreMsgServer.PostDataRequest(
		f.Context(),
		&types.MsgPostDataRequest{
			Sender:            f.Creator.Address(),
			Funds:             sdk.NewCoin(BondDenom, attachedFunds),
			Version:           "0.0.1",
			ExecProgramID:     hex.EncodeToString(dr.ExecProgHash),
			ExecInputs:        []byte("exec_inputs"),
			ExecGasLimit:      execGasLimit,
			TallyProgramID:    hex.EncodeToString(dr.TallyProgHash),
			TallyInputs:       []byte("tally_inputs"),
			TallyGasLimit:     tallyGasLimit,
			ReplicationFactor: uint32(dr.ReplicationFactor), //nolint:gosec
			ConsensusFilter:   []byte{0x00},
			GasPrice:          math.NewInt(2000),
			Memo:              []byte(dr.Memo),
			SEDAPayload:       []byte{},
			PaybackAddress:    []byte{0x01, 0x02, 0x03},
		},
	)
	require.NoError(f.tb, err)

	dr.postedID = res.DrID
	dr.postedHeight = uint64(res.Height) //nolint:gosec
}

// CommitDataRequest executes commit messages from the first n stakers.
// If stakerIndices is not nil, stakers at the specified indices will be used.
func (dr *TestDR) CommitDataRequest(f *Fixture, n int, stakerIndices []int) {
	if stakerIndices != nil {
		require.Equal(f.tb, n, len(stakerIndices))
	}

	gasLimit := dr.CommitGasLimit
	if dr.CommitGasLimit == 0 {
		// TODO Use gastimation formula
		gasLimit = 500000
	}

	revealBody := types.RevealBody{
		DrID:         dr.postedID,
		Reveal:       dr.Reveal,
		GasUsed:      dr.GasUsed,
		ExitCode:     uint32(dr.ExitCode),
		ProxyPubKeys: dr.ProxyPubKeys,
	}

	var indices []int
	if stakerIndices != nil {
		indices = stakerIndices
	} else {
		// Use indices 0, 1, ..., n-1.
		indices = make([]int, n)
		for i := 0; i < n; i++ {
			indices[i] = i
		}
	}

	for _, i := range indices {
		revealMsg, commitment, revealProof := f.createRevealMsg(f.Stakers[i], revealBody)
		commitMsg := &types.MsgCommit{
			Sender:    f.Stakers[i].Address.String(),
			DrID:      dr.postedID,
			Commit:    commitment,
			PublicKey: f.Stakers[i].PubKey,
		}
		commitMsg.Proof = f.Stakers[i].GenerateProof(f.tb, commitMsg.MsgHash("", f.ChainID, f.Context().BlockHeight()))

		f.executeMsg(commitMsg, f.Stakers[i].Address, gasLimit)

		dr.revealMsgs = append(dr.revealMsgs, revealMsg)
		dr.revealProofs = append(dr.revealProofs, revealProof)
	}
}

// RevealDataRequest executes reveal messages stored from execution of
// CommitDataRequest using the first n stakers.
// If stakerIndices is not nil, stakers at the specified indices will be used.
// After execution, stored reveal messages are cleared.
func (dr *TestDR) RevealDataRequest(f *Fixture, n int, stakerIndices []int) {
	require.LessOrEqual(f.tb, n, len(dr.revealMsgs))
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
		// Use indices 0, 1, ..., n-1.
		indices = make([]int, n)
		for i := 0; i < n; i++ {
			indices[i] = i
		}
	}

	for _, i := range indices {
		f.executeMsg(&dr.revealMsgs[i], f.Stakers[i].Address, gasLimit)
	}
	dr.revealMsgs = nil
}
