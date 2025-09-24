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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	RequestHeight  uint64
	RequestMemo    string
	Reveal         []byte
	ProxyPubKeys   []string
	GasUsed        uint64
	ExitCode       byte
	CommitGasLimit uint64 // defaults to 100000 if not set or set to 0
	RevealGasLimit uint64 // defaults to 100000 if not set or set to 0
}

func (f *Fixture) uploadOraclePrograms(tb testing.TB) {
	tb.Helper()

	for _, op := range testwasms.TestWasms {
		execProgram := wasmstoragetypes.NewOracleProgram(op, f.Context().BlockTime())
		err := f.WasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
		require.NoError(tb, err)
	}
}

// ExecuteDataRequestFlow posts a data request using the deployer account and
// executes given numbers of commits and reveals for the request using the
// stakers accounts. It returns the data request ID.
func (f *Fixture) ExecuteDataRequestFlow(
	execProgramBytes, tallyProgramBytes []byte,
	replicationFactor, numCommits, numReveals int, timeout bool,
	config CommitRevealConfig,
) string {
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
	res := f.PostDataRequest(execProgram.Hash, tallyProgram.Hash, config.RequestMemo, replicationFactor)

	drID := res.DrID

	// The stakers commit and reveal.
	revealMsgs := f.CommitDataRequest(f.Stakers[:numCommits], res.Height, drID, config)

	f.ExecuteReveals(f.Stakers[:numReveals], revealMsgs[:numReveals], config)

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
// func (f *Fixture) executeDataRequestFlowWithTallyTestItem(tb testing.TB, entropy []byte) (string, testwasms.TallyTestItem) {
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
// 	res := f.PostDataRequest(tb, execProgram.Hash, tallyProgram.Hash, config.RequestMemo, 1)
// 	drID := res.DrID

// 	// The stakers commit and reveal.
// 	revealMsgs := f.CommitDataRequest(tb, f.Stakers[:1], res.Height, drID, config)
// 	f.ExecuteReveals(tb, f.Stakers, revealMsgs[:1], config)

// 	return res.DrID, testItem
// }

func (f *Fixture) PostDataRequest(execProgHash, tallyProgHash []byte, requestMemo string, replicationFactor int) PostDataRequestResponse {
	amount, ok := math.NewIntFromString("200600000000000000000")
	require.True(f.tb, ok)

	resJSON := f.executeCoreContract(
		f.Creator.Address(),
		testutil.PostDataRequestMsg(execProgHash, tallyProgHash, requestMemo, replicationFactor),
		sdk.NewCoins(sdk.NewCoin(BondDenom, amount)),
	)

	var res PostDataRequestResponse
	err := json.Unmarshal(resJSON, &res)
	require.NoError(f.tb, err)
	return res
}

// commitDataRequest executes a commit for each of the given stakers and
// returns a list of corresponding reveal messages.
func (f *Fixture) CommitDataRequest(stakers []Staker, height uint64, drID string, config CommitRevealConfig) [][]byte {
	CommitGasLimit := config.CommitGasLimit
	if config.CommitGasLimit == 0 {
		CommitGasLimit = 100000
	}

	revealBody := types.RevealBody{
		DrID:         drID,
		Reveal:       config.Reveal,
		GasUsed:      config.GasUsed,
		ExitCode:     uint32(config.ExitCode),
		ProxyPubKeys: config.ProxyPubKeys,
	}

	var revealMsgs [][]byte
	for i := 0; i < len(stakers); i++ {
		revealMsg, commitment, _ := f.createRevealMsg(stakers[i], revealBody)

		proof := f.generateCommitProof(stakers[i].Key, drID, commitment, height)

		commitMsg := testutil.CommitMsg(drID, commitment, stakers[i].PubKey, proof)
		f.executeCommitReveal(stakers[i].Address, commitMsg, CommitGasLimit)

		revealMsgs = append(revealMsgs, revealMsg)
	}

	return revealMsgs
}

// executeReveals executes a list of reveal messages, one by one using
// the staker addresses.
func (f *Fixture) ExecuteReveals(stakers []Staker, revealMsgs [][]byte, config CommitRevealConfig) {
	RevealGasLimit := config.RevealGasLimit
	if config.RevealGasLimit == 0 {
		RevealGasLimit = 100000
	}

	require.Equal(f.tb, len(stakers), len(revealMsgs))

	for i := range revealMsgs {
		f.executeCommitReveal(stakers[i].Address, revealMsgs[i], RevealGasLimit)
	}
}

type Staker struct {
	Key     []byte
	PubKey  string
	Address []byte
}

func (f *Fixture) mintCoinsForAccount(tb testing.TB, address sdk.AccAddress, sedaAmount uint64) {
	tb.Helper()

	amount := math.NewIntFromUint64(sedaAmount).Mul(math.NewInt(1e18))
	err := f.BankKeeper.MintCoins(f.Context(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(BondDenom, amount)))
	require.NoError(tb, err)
	err = f.BankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, address, sdk.NewCoins(sdk.NewCoin(BondDenom, amount)))
	require.NoError(tb, err)
}

// AddStakers generates stakers and adds them to the allowlist. The
// stakers subsequently send their stakes to the core contract.
func (f *Fixture) AddStakers(tb testing.TB, num int) []Staker {
	tb.Helper()

	stakers := make([]Staker, num)
	for i := 0; i < num; i++ {
		privKey := secp256k1.GenPrivKey()
		stakers[i] = Staker{
			Key:     privKey.Bytes(),
			PubKey:  hex.EncodeToString(privKey.PubKey().Bytes()),
			Address: privKey.PubKey().Address().Bytes(),
		}

		f.mintCoinsForAccount(tb, stakers[i].Address, 100)

		// Add to allowlist.
		f.executeCoreContract(
			f.Creator.Address(),
			testutil.AddToAllowListMsg(stakers[i].PubKey),
			sdk.NewCoins(),
		)

		// Stake.
		f.initAccountWithCoins(tb, stakers[i].Address, sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1e18))))
		proof := f.generateStakeProof(tb, stakers[i].Key, "YWRkcmVzcw==", 0)
		f.executeCoreContract(
			sdk.AccAddress(stakers[i].Address).String(),
			testutil.StakeMsg(stakers[i].PubKey, proof, "YWRkcmVzcw=="),
			sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1000000000000000000))),
		)

		// Second stake to test sequence number.
		proof = f.generateStakeProof(tb, stakers[i].Key, "YWRkcmVzcw==", 1)
		f.executeCoreContract(
			f.Creator.Address(),
			testutil.StakeMsg(stakers[i].PubKey, proof, "YWRkcmVzcw=="),
			sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(500000000000000000))),
		)
	}

	f.Stakers = append(f.Stakers, stakers...)
	return stakers
}

// CheckDataRequestsByStatus checks that the given number of data requests is
// retrieved by GetDataRequestsByStatus.
func (f *Fixture) CheckDataRequestsByStatus(tb testing.TB, status types.DataRequestStatus, expectedTotal, fetchLimit uint64) {
	tb.Helper()

	remainingTotal := expectedTotal

	drs, lastSeen, total, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), status, fetchLimit, nil)
	require.NoError(tb, err)
	require.Equal(tb, expectedTotal, total)
	require.Equal(tb, min(remainingTotal, fetchLimit), uint64(len(drs)))

	for fetchLimit < remainingTotal {
		remainingTotal -= fetchLimit

		require.NotNil(tb, lastSeen)
		drs, lastSeen, total, err = f.CoreKeeper.GetDataRequestsByStatus(f.Context(), status, fetchLimit, lastSeen)
		require.NoError(tb, err)
		require.Equal(tb, expectedTotal, total)
		require.Equal(tb, min(remainingTotal, fetchLimit), uint64(len(drs)))
	}
}

// func (f *Fixture) pauseContract(tb testing.TB) {
// 	f.executeCoreContract(tb, f.deployer.String(), []byte(`{"pause":{}}`), sdk.NewCoins())
// }

// generateStakeProof generates a proof for a stake message given a
// base64-encoded memo.
func (f *Fixture) generateStakeProof(tb testing.TB, signKey []byte, base64Memo string, seqNum uint64) string {
	tb.Helper()

	msg := types.MsgStake{
		Memo: base64Memo,
	}
	hash, err := msg.MsgHash(f.ChainID, seqNum)
	require.NoError(tb, err)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(tb, err)
	return hex.EncodeToString(proof)
}

func (f *Fixture) generateCommitProof(signKey []byte, drID, commitment string, drHeight uint64) string {
	f.tb.Helper()

	commitBytes := []byte("commit_data_result")
	drIDBytes := []byte(drID)

	drHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(drHeightBytes, drHeight)

	commitmentBytes := []byte(commitment)
	chainIDBytes := []byte(f.ChainID)

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
	require.NoError(f.tb, err)

	return hex.EncodeToString(proof)
}

func (f *Fixture) initAccountWithCoins(tb testing.TB, addr sdk.AccAddress, coins sdk.Coins) {
	tb.Helper()

	err := f.BankKeeper.MintCoins(f.Context(), minttypes.ModuleName, coins)
	require.NoError(tb, err)
	err = f.BankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, addr, coins)
	require.NoError(tb, err)
}

// createRevealMsg constructs and returns a reveal message and its corresponding
// commitment and proof.
func (f *Fixture) createRevealMsg(staker Staker, revealBody types.RevealBody) ([]byte, string, string) {
	f.tb.Helper()

	revealBodyHash, err := revealBody.RevealBodyHash()
	require.NoError(f.tb, err)

	proof := generateRevealProof(f.tb, staker.Key, revealBodyHash, f.ChainID, f.CoreContractAddr.String())

	msg := testutil.RevealMsg(
		revealBody.DrID,
		base64.StdEncoding.EncodeToString(revealBody.Reveal),
		staker.PubKey,
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
	hasher.Write([]byte(staker.PubKey))
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
func (f *Fixture) executeCommitReveal(sender sdk.AccAddress, msg []byte, gasLimit uint64) {
	f.SetBasicGasMeter(gasLimit)

	contractMsg := wasmtypes.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: f.CoreContractAddr.String(),
		Msg:      msg,
		Funds:    sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1))),
	}

	fee := sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(gasLimit).Mul(math.NewInt(1e10))))
	txf := tx.Factory{}.
		WithChainID(f.ChainID).
		WithTxConfig(f.TxConfig).
		WithFees(fee.String()).
		WithFeePayer(sender)

	tx, err := txf.BuildUnsignedTx(&contractMsg)
	require.NoError(f.tb, err)

	txBytes, err := f.TxConfig.TxEncoder()(tx.GetTx())
	require.NoError(f.tb, err)

	f.SetContextTxBytes(txBytes)

	// Transfer the fee to the fee collector.
	// This simulates the ante handler DeductFees.
	err = f.BankKeeper.SendCoinsFromAccountToModule(f.Context(), sender, authtypes.FeeCollectorName, fee)
	require.NoError(f.tb, err)

	// Execute the message.
	f.executeCoreContract(sender.String(), msg, sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1))))

	f.SetInfiniteGasMeter()
}

func (f *Fixture) executeCoreContract(sender string, msg []byte, funds sdk.Coins) []byte {
	execMsg := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: f.CoreContractAddr.String(),
		Msg:      msg,
		Funds:    funds,
	}

	handler := f.Router.Handler(execMsg)
	require.NotNil(f.tb, handler)

	result, err := handler(f.Context(), execMsg)
	require.NoError(f.tb, err, "failed to execute Core Contract msg %s", execMsg.String())

	return result.MsgResponses[0].GetCachedValue().(*wasmtypes.MsgExecuteContractResponse).Data
}
