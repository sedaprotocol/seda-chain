package testutil

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
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

func (dr *TestDR) PostDataRequestShouldErr(f *Fixture, errMsg string) {
	amount, ok := math.NewIntFromString("200600000000000000000")
	require.True(f.tb, ok)

	f.executeCoreContractShouldErr(
		f.Creator.Address(),
		testutil.PostDataRequestMsg(dr.ExecProgHash, dr.TallyProgHash, dr.Memo, dr.ReplicationFactor),
		sdk.NewCoins(sdk.NewCoin(BondDenom, amount)),
		errMsg,
	)
}

// commitDataRequest executes a commit for each of the given stakers and
// returns a list of corresponding reveal messages.
func (dr *TestDR) CommitDataRequest(f *Fixture, numCommits int) {
	require.LessOrEqual(f.tb, numCommits, len(f.Stakers))

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
		gasLimit = 500000
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

func (ta *TestAccount) GetDataRequestConfig() (*types.QueryDataRequestConfigResponse, error) {
	msg := &types.QueryDataRequestConfigRequest{}
	return ta.fixture.CoreQuerier.DataRequestConfig(ta.fixture.Context(), msg)
}

func HashStringHelper(input string) []byte {
	return crypto.Keccak256([]byte(input))
}

func RevealHelperFromString(input string) []byte {
	return []byte(base64.StdEncoding.EncodeToString(HashStringHelper(input)))
}

func (ta *TestAccount) CalculateDrIdAndArgs(nonce string, replication_factor uint32) types.MsgPostDataRequest {
	execProgramId := hex.EncodeToString(HashStringHelper(nonce))
	execInputs := base64.StdEncoding.EncodeToString(HashStringHelper("exec_inputs"))
	tallyProgramId := hex.EncodeToString(HashStringHelper("tally_program"))
	tallyInputs := base64.StdEncoding.EncodeToString(HashStringHelper("tally_inputs"))

	memo := base64.StdEncoding.EncodeToString(crypto.Keccak256([]byte(ta.fixture.ChainID), []byte(nonce)))

	return types.MsgPostDataRequest{
		Sender:            ta.Address(),
		Version:           "1.0.0",
		ExecProgramID:     execProgramId,
		ExecInputs:        []byte(execInputs),
		ExecGasLimit:      types.MinExecGasLimit,
		TallyProgramID:    tallyProgramId,
		TallyInputs:       []byte(tallyInputs),
		TallyGasLimit:     types.MinTallyGasLimit,
		Memo:              []byte(memo),
		ReplicationFactor: replication_factor,
		ConsensusFilter:   []byte(base64.StdEncoding.EncodeToString([]byte{0})),
		GasPrice:          types.MinGasPrice,
	}
}

func MinimumDrCost() sdk.Coin {
	return sdk.NewCoin(BondDenom, (math.NewIntFromUint64(types.MinExecGasLimit).Add(math.NewIntFromUint64(types.MinTallyGasLimit)).Mul(types.MinGasPrice)))
}

func (ta *TestAccount) PostDataRequest(args types.MsgPostDataRequest, blockHeight int64, funds *math.Int) (*types.MsgPostDataRequestResponse, error) {
	if blockHeight < ta.fixture.Context().BlockHeight() {
		panic("cannot set block height to the past")
	}

	ta.fixture.Context().WithBlockHeight(blockHeight)

	if funds != nil {
		args.Funds = sdk.NewCoin(BondDenom, *funds)
	} else {
		args.Funds = MinimumDrCost()
	}

	return ta.fixture.CoreMsgServer.PostDataRequest(ta.fixture.Context(), &args)
}

func (ta *TestAccount) CreateRevealMsg(revealBody *types.RevealBody) *types.MsgReveal {
	msg := &types.MsgReveal{
		Sender:     ta.Address(),
		RevealBody: revealBody,
		PublicKey:  ta.PublicKeyHex(),
		Stderr:     []string{},
		Stdout:     []string{},
	}
	hash, err := msg.MsgHash(ta.fixture.ChainID)
	require.NoError(ta.fixture.tb, err)
	msg.Proof = ta.Prove(hash)
	return msg
}

func (ta *TestAccount) CommitResult(revealMsg *types.MsgReveal) (*types.MsgCommitResponse, error) {
	commitment, err := revealMsg.RevealHash()
	require.NoError(ta.fixture.tb, err)

	msg := &types.MsgCommit{
		Sender:    ta.Address(),
		DrID:      revealMsg.RevealBody.DrID,
		Commit:    hex.EncodeToString(commitment),
		PublicKey: ta.PublicKeyHex(),
	}
	hash, err := msg.MsgHash(ta.fixture.ChainID, int64(revealMsg.RevealBody.DrBlockHeight))
	require.NoError(ta.fixture.tb, err)
	msg.Proof = ta.Prove(hash)

	ta.fixture.SetTx(20_000, ta.AccAddress(), msg)
	res, err := ta.fixture.CoreMsgServer.Commit(ta.fixture.Context(), msg)
	ta.fixture.SetInfiniteGasMeter()
	return res, err
}

func (ta *TestAccount) RevealResult(msg *types.MsgReveal) (*types.MsgRevealResponse, error) {
	ta.fixture.SetTx(20_000, ta.AccAddress(), msg)
	res, err := ta.fixture.CoreMsgServer.Reveal(ta.fixture.Context(), msg)
	ta.fixture.SetInfiniteGasMeter()
	return res, err
}

func (ta *TestAccount) GetDataRequestsByStatus(status types.DataRequestStatus, limit uint64, lastSeenIndex *[]string) (*types.QueryDataRequestsByStatusResponse, error) {
	msg := &types.QueryDataRequestsByStatusRequest{
		Status: status,
		Limit:  limit,
	}
	if lastSeenIndex != nil {
		msg.LastSeenIndex = *lastSeenIndex
	}

	return ta.fixture.CoreQuerier.DataRequestsByStatus(ta.fixture.Context(), msg)
}

func (ta *TestAccount) GetDataRequest(drId string) (*types.QueryDataRequestResponse, error) {
	msg := &types.QueryDataRequestRequest{
		DrId: drId,
	}
	return ta.fixture.CoreQuerier.DataRequest(ta.fixture.Context(), msg)
}
