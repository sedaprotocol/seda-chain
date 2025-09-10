package testutil

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/rand"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	vrf "github.com/sedaprotocol/vrf-go"

	"github.com/sedaprotocol/seda-chain/testutil"
	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/core/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	defaultCommitTimeoutBlocks = 50
	defaultRevealTimeoutBlocks = 5
)

type PostDataRequestResponse struct {
	DrID   string `json:"dr_id"`
	Height uint64 `json:"height"`
}

type CommitRevealConfig struct {
	RequestHeight uint64
	RequestMemo   string
	Reveal        []byte
	ProxyPubKeys  []string
	GasUsed       uint64
	ExitCode      byte
}

func (f *Fixture) uploadOraclePrograms(tb testing.TB) {
	tb.Helper()

	for _, op := range testwasms.TestWasms {
		execProgram := wasmstoragetypes.NewOracleProgram(op, f.Context().BlockTime())
		err := f.wasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
		require.NoError(tb, err)
	}
}

// ExecuteDataRequestFlow posts a data request using the deployer account and
// executes given numbers of commits and reveals for the request using the
// stakers accounts. It returns the data request ID.
func (f *Fixture) ExecuteDataRequestFlow(
	tb testing.TB,
	execProgramBytes, tallyProgramBytes []byte,
	replicationFactor, numCommits, numReveals int, timeout bool,
	config CommitRevealConfig,
) string {
	tb.Helper()

	var execProgram, tallyProgram wasmstoragetypes.OracleProgram
	if execProgramBytes != nil {
		execProgram = wasmstoragetypes.NewOracleProgram(execProgramBytes, f.Context().BlockTime())
	} else {
		randIndex := rand.Intn(len(testwasms.TestWasms))
		execProgram = wasmstoragetypes.NewOracleProgram(testwasms.TestWasms[randIndex], f.Context().BlockTime())
	}
	if tallyProgramBytes != nil {
		tallyProgram = wasmstoragetypes.NewOracleProgram(tallyProgramBytes, f.Context().BlockTime())
	} else {
		randIndex := rand.Intn(len(testwasms.TestWasms))
		tallyProgram = wasmstoragetypes.NewOracleProgram(testwasms.TestWasms[randIndex], f.Context().BlockTime())
	}

	// Post a data request.
	res := f.postDataRequest(tb, execProgram.Hash, tallyProgram.Hash, config.RequestMemo, replicationFactor)

	drID := res.DrID

	// The stakers commit and reveal.
	revealMsgs := f.commitDataRequest(tb, f.stakers[:numCommits], res.Height, drID, config)

	f.executeReveals(tb, f.stakers, revealMsgs[:numReveals])

	if timeout {
		timeoutBlocks := defaultCommitTimeoutBlocks
		if numCommits == replicationFactor {
			timeoutBlocks = defaultRevealTimeoutBlocks
		}

		for range timeoutBlocks {
			f.AddBlock()
		}
	}
	return res.DrID
}

// executeDataRequestFlowWithTallyTestItem posts a data request using the deployer
// account and executes a commit and reveal for the request using a staker account.
// It then returns the data request ID and the randomly selected TallyTestItem,
// which contains the expected tally execution results.
// func (f *fixture) executeDataRequestFlowWithTallyTestItem(tb testing.TB, entropy []byte) (string, testwasms.TallyTestItem) {
// 	randIndex := rand.Intn(len(testwasms.TestWasms))
// 	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.TestWasms[randIndex], f.Context().BlockTime())

// 	randIndex = rand.Intn(len(testwasms.TallyTestItems))
// 	testItem := testwasms.TallyTestItems[randIndex]
// 	tallyProgram := wasmstoragetypes.NewOracleProgram(testItem.TallyProgram, f.Context().BlockTime())

// 	config := CommitRevealConfig{
// 		RequestHeight: 1,
// 		RequestMemo:   base64.StdEncoding.EncodeToString(entropy),
// 		Reveal:        testItem.Reveal,
// 		ProxyPubKeys:  []string{},
// 		GasUsed:       testItem.GasUsed,
// 	}

// 	// Post a data request.
// 	res, err := f.postDataRequest(tb, execProgram.Hash, tallyProgram.Hash, config.RequestMemo, 1)
// 	require.NoError(tb, err)

// 	drID := res.DrID

// 	// The stakers commit and reveal.
// 	revealMsgs, err := f.commitDataRequest(tb, f.stakers[:1], res.Height, drID, config)
// 	require.NoError(tb, err)

// 	err = f.executeReveals(tb, f.stakers, revealMsgs[:1])
// 	require.NoError(tb, err)

// 	return res.DrID, testItem
// }

func (f *Fixture) postDataRequest(tb testing.TB, execProgHash, tallyProgHash []byte, requestMemo string, replicationFactor int) PostDataRequestResponse {
	tb.Helper()

	amount, ok := math.NewIntFromString("200600000000000000000")
	require.True(tb, ok)

	resJSON := f.executeCoreContract(
		tb, f.Creator.Address(),
		testutil.PostDataRequestMsg(execProgHash, tallyProgHash, requestMemo, replicationFactor),
		sdk.NewCoins(sdk.NewCoin(bondDenom, amount)),
	)

	var res PostDataRequestResponse
	err := json.Unmarshal(resJSON, &res)
	require.NoError(tb, err)
	return res
}

// commitDataRequest executes a commit for each of the given stakers and
// returns a list of corresponding reveal messages.
func (f *Fixture) commitDataRequest(tb testing.TB, stakers []staker, height uint64, drID string, config CommitRevealConfig) [][]byte {
	tb.Helper()

	revealBody := types.RevealBody{
		DrID:         drID,
		Reveal:       config.Reveal,
		GasUsed:      config.GasUsed,
		ExitCode:     uint32(config.ExitCode),
		ProxyPubKeys: config.ProxyPubKeys,
	}

	var revealMsgs [][]byte
	for i := 0; i < len(stakers); i++ {
		revealMsg, commitment, _ := f.createRevealMsg(tb, stakers[i], revealBody)

		proof := f.generateCommitProof(tb, stakers[i].key, drID, commitment, height)

		commitMsg := testutil.CommitMsg(drID, commitment, stakers[i].pubKey, proof)
		f.executeCommitReveal(tb, stakers[i].address, commitMsg, 500000)

		revealMsgs = append(revealMsgs, revealMsg)
	}

	return revealMsgs
}

// executeReveals executes a list of reveal messages.
func (f *Fixture) executeReveals(tb testing.TB, stakers []staker, revealMsgs [][]byte) {
	tb.Helper()

	for i := 0; i < len(revealMsgs); i++ {
		f.executeCommitReveal(tb, stakers[i].address, revealMsgs[i], 500000)
	}
}

type staker struct {
	key     []byte
	pubKey  string
	address []byte
}

// addStakers generates stakers and adds them to the allowlist. The
// stakers subsequently send their stakes to the core contract.
func (f *Fixture) addStakers(tb testing.TB, num int) []staker {
	tb.Helper()

	stakers := make([]staker, num)
	for i := 0; i < num; i++ {
		privKey := secp256k1.GenPrivKey()
		stakers[i] = staker{
			key:     privKey.Bytes(),
			pubKey:  hex.EncodeToString(privKey.PubKey().Bytes()),
			address: privKey.PubKey().Address().Bytes(),
		}

		// Add to allowlist.
		f.executeCoreContract(
			tb, f.Creator.Address(),
			testutil.AddToAllowListMsg(stakers[i].pubKey),
			sdk.NewCoins(),
		)

		// Stake.
		f.initAccountWithCoins(tb, stakers[i].address, sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1e18))))
		proof := f.generateStakeProof(tb, stakers[i].key, "YWRkcmVzcw==", 0)
		f.executeCoreContract(
			tb, sdk.AccAddress(stakers[i].address).String(),
			testutil.StakeMsg(stakers[i].pubKey, proof, "YWRkcmVzcw=="),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1000000000000000000))),
		)

		// Second stake to test sequence number.
		proof = f.generateStakeProof(tb, stakers[i].key, "YWRkcmVzcw==", 1)
		f.executeCoreContract(
			tb, f.Creator.Address(),
			testutil.StakeMsg(stakers[i].pubKey, proof, "YWRkcmVzcw=="),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(500000000000000000))),
		)
	}

	f.stakers = stakers
	return stakers
}

// func (f *fixture) pauseContract(tb testing.TB) {
// 	f.executeCoreContract(tb, f.deployer.String(), []byte(`{"pause":{}}`), sdk.NewCoins())
// }

// generateStakeProof generates a proof for a stake message given a
// base64-encoded memo.
func (f *Fixture) generateStakeProof(tb testing.TB, signKey []byte, base64Memo string, seqNum uint64) string {
	tb.Helper()

	msg := types.MsgStake{
		Memo: base64Memo,
	}
	hash, err := msg.MsgHash(f.chainID, seqNum)
	require.NoError(tb, err)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(tb, err)
	return hex.EncodeToString(proof)
}

func (f *Fixture) generateCommitProof(tb testing.TB, signKey []byte, drID, commitment string, drHeight uint64) string {
	tb.Helper()

	commitBytes := []byte("commit_data_result")
	drIDBytes := []byte(drID)

	drHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(drHeightBytes, drHeight)

	commitmentBytes := []byte(commitment)
	chainIDBytes := []byte(f.chainID)

	allBytes := append([]byte{}, commitBytes...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, commitmentBytes...)
	allBytes = append(allBytes, chainIDBytes...)
	// Legacy format
	// allBytes = append(allBytes, []byte(f.coreContractAddr.String())...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(tb, err)

	return hex.EncodeToString(proof)
}

func (f *Fixture) initAccountWithCoins(tb testing.TB, addr sdk.AccAddress, coins sdk.Coins) {
	tb.Helper()

	err := f.bankKeeper.MintCoins(f.Context(), minttypes.ModuleName, coins)
	require.NoError(tb, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, addr, coins)
	require.NoError(tb, err)
}

// createRevealMsg constructs and returns a reveal message and its corresponding
// commitment and proof.
func (f *Fixture) createRevealMsg(tb testing.TB, staker staker, revealBody types.RevealBody) ([]byte, string, string) {
	tb.Helper()

	revealBodyHash, err := revealBody.RevealBodyHash()
	require.NoError(tb, err)

	proof := generateRevealProof(tb, staker.key, revealBodyHash, f.chainID, f.coreContractAddr.String())

	msg := testutil.RevealMsg(
		revealBody.DrID,
		base64.StdEncoding.EncodeToString(revealBody.Reveal),
		staker.pubKey,
		proof,
		revealBody.ProxyPubKeys,
		byte(revealBody.ExitCode),
		revealBody.DrBlockHeight,
		revealBody.GasUsed,
	)

	// commitment = hash(revealBodyHash | publicKey | proof | stderr | stdout)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("reveal_message"))
	hasher.Write(revealBodyHash)
	hasher.Write([]byte(staker.pubKey))
	hasher.Write([]byte(proof))
	hasher.Write([]byte(strings.Join([]string{""}, "")))
	hasher.Write([]byte(strings.Join([]string{""}, "")))
	commitment := hasher.Sum(nil)

	return msg, hex.EncodeToString(commitment), proof
}

func generateRevealProof(tb testing.TB, signKey []byte, revealBodyHash []byte, chainID, _ string) string {
	tb.Helper()

	allBytes := []byte("reveal_data_result")

	allBytes = append(allBytes, revealBodyHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	// Legacy format
	// allBytes = append(allBytes, []byte(coreContractAddr)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(tb, err)

	return hex.EncodeToString(proof)
}

// executeCommitReveal executes a commit msg or a reveal msg with the required
// context.
func (f *Fixture) executeCommitReveal(tb testing.TB, sender sdk.AccAddress, msg []byte, gasLimit uint64) {
	tb.Helper()

	contractMsg := wasmtypes.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: f.coreContractAddr.String(),
		Msg:      msg,
		Funds:    sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
	}

	fee := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(gasLimit*1e10)))
	txf := tx.Factory{}.
		WithChainID(f.chainID).
		WithTxConfig(f.txConfig).
		WithFees(fee.String()).
		WithFeePayer(sender)

	tx, err := txf.BuildUnsignedTx(&contractMsg)
	require.NoError(tb, err)

	txBytes, err := f.txConfig.TxEncoder()(tx.GetTx())
	require.NoError(tb, err)

	f.SetContextTxBytes(txBytes)
	f.SetBasicGasMeter(gasLimit)

	// TODO reactivate once refund is implemented
	// Transfer the fee to the fee collector.
	// err = f.bankKeeper.SendCoinsFromAccountToModule(f.Context(), sender, authtypes.FeeCollectorName, fee)
	// require.NoError(t, err)

	// Execute the message.
	f.executeCoreContract(tb, sender.String(), msg, sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))))

	f.SetInfiniteGasMeter()
}

func (f *Fixture) executeCoreContract(tb testing.TB, sender string, msg []byte, funds sdk.Coins) []byte {
	tb.Helper()

	execMsg := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: f.coreContractAddr.String(),
		Msg:      msg,
		Funds:    funds,
	}

	handler := f.router.Handler(execMsg)
	require.NotNil(tb, handler)

	result, err := handler(f.Context(), execMsg)
	require.NoError(tb, err, "failed to execute Core Contract msg %s", execMsg.String())

	return result.MsgResponses[0].GetCachedValue().(*wasmtypes.MsgExecuteContractResponse).Data
}
