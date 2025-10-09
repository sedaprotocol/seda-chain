package testutil

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

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

func (f *Fixture) uploadOraclePrograms(tb testing.TB) {
	tb.Helper()

	for _, op := range testwasms.TestWasms {
		execProgram := wasmstoragetypes.NewOracleProgram(op, f.Context().BlockTime())
		err := f.WasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
		require.NoError(tb, err)
	}
}

func (f *Fixture) mintCoinsForAccount(tb testing.TB, address sdk.AccAddress, sedaAmount uint64) {
	tb.Helper()

	amount := math.NewIntFromUint64(sedaAmount).Mul(math.NewInt(1e18))
	err := f.BankKeeper.MintCoins(f.Context(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(BondDenom, amount)))
	require.NoError(tb, err)
	err = f.BankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, address, sdk.NewCoins(sdk.NewCoin(BondDenom, amount)))
	require.NoError(tb, err)
}

type Staker struct {
	Key     []byte
	PubKey  string
	Address []byte
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

func (f *Fixture) DrainDataRequestPool(targetHeight uint64) []byte {
	return f.executeCoreContract(
		f.Creator.Address(),
		testutil.DrainDataRequestPoolMsg(targetHeight),
		sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(500000000000000000))),
	)
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

// generateStakeProof generates a proof for a stake message given a
// base64-encoded memo.
func (f *Fixture) generateStakeProof(tb testing.TB, signKey []byte, base64Memo string, seqNum uint64) string {
	tb.Helper()

	var hash []byte
	var err error
	if f.noShim {
		msg := types.MsgLegacyStake{
			Memo: base64Memo,
		}
		hash, err = msg.MsgHash(f.CoreContractAddr.String(), f.ChainID, seqNum)
	} else {
		msg := types.MsgStake{
			Memo: base64Memo,
		}
		hash, err = msg.MsgHash(f.ChainID, seqNum)
	}
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
	if f.noShim {
		allBytes = append(allBytes, []byte(f.CoreContractAddr.String())...)
	}

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

	proof := f.generateRevealProof(f.tb, staker.Key, revealBodyHash)

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

func (f *Fixture) generateRevealProof(tb testing.TB, signKey []byte, revealBodyHash []byte) string {
	tb.Helper()

	allBytes := []byte("reveal_data_result")
	allBytes = append(allBytes, revealBodyHash...)
	allBytes = append(allBytes, []byte(f.ChainID)...)
	if f.noShim {
		allBytes = append(allBytes, []byte(f.CoreContractAddr.String())...)
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(tb, err)

	return hex.EncodeToString(proof)
}

// executeCommitOrReveal executes a commit msg or a reveal msg.
func (f *Fixture) executeCommitOrReveal(sender sdk.AccAddress, msg []byte, gasLimit uint64) {
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

	f.SetBasicGasMeter(gasLimit)

	// Transfer the fee to the fee collector.
	// This simulates the ante handler DeductFees.
	err = f.BankKeeper.SendCoinsFromAccountToModule(f.Context(), sender, authtypes.FeeCollectorName, fee)
	require.NoError(f.tb, err)

	// Execute the message.
	f.executeCoreContract(sender.String(), msg, sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1))))

	// Reset to infinite gas meter to prevent out of gas error in other operations.
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

func (f *Fixture) executeCoreContractShouldErr(sender string, msg []byte, funds sdk.Coins, errMsg string) {
	execMsg := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: f.CoreContractAddr.String(),
		Msg:      msg,
		Funds:    funds,
	}

	handler := f.Router.Handler(execMsg)
	require.NotNil(f.tb, handler)

	_, err := handler(f.Context(), execMsg)
	require.Error(f.tb, err)
	require.Contains(f.tb, err.Error(), errMsg)
}
