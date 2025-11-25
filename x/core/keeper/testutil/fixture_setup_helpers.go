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

func (f *Fixture) executeMsg(msg sdk.Msg, sender sdk.AccAddress, gasLimit uint64) *sdk.Result {
	fee := sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(gasLimit).Mul(math.NewInt(1e10))))
	txf := tx.Factory{}.
		WithChainID(f.ChainID).
		WithTxConfig(f.TxConfig).
		WithFees(fee.String()).
		WithFeePayer(sender)

	tx, err := txf.BuildUnsignedTx(msg)
	require.NoError(f.tb, err)
	txBytes, err := f.TxConfig.TxEncoder()(tx.GetTx())
	require.NoError(f.tb, err)
	f.SetContextTxBytes(txBytes)

	if gasLimit > 0 {
		f.SetBasicGasMeter(gasLimit)
	} else {
		f.SetInfiniteGasMeter()
	}

	// Transfer the fee to the fee collector.
	// This simulates the ante handler DeductFees.
	err = f.BankKeeper.SendCoinsFromAccountToModule(f.Context(), sender, authtypes.FeeCollectorName, fee)
	require.NoError(f.tb, err)

	handler := f.Router.Handler(msg)
	require.NotNil(f.tb, handler)

	result, err := handler(f.Context(), msg)
	require.NoError(f.tb, err)

	// Reset to infinite gas meter to prevent out of gas error in other operations.
	f.SetInfiniteGasMeter()

	return result
}

func (f *Fixture) uploadOraclePrograms(tb testing.TB) {
	tb.Helper()

	for i, op := range testwasms.TestWasms {
		execProgram := wasmstoragetypes.NewOracleProgram(op, f.Context().BlockTime())
		err := f.WasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
		f.DeployedOPs[testwasms.TestWasmNames[i]] = hex.EncodeToString(execProgram.Hash)
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
	PrivKey []byte
	PubKey  string
	// Address is the SEDA account address that holds SEDA tokens. It is normally
	// separate from staker (privKey, pubKey) pair, but we derive the address from
	// the staker private key here for convenience.
	Address sdk.AccAddress
}

type MsgHasher interface {
	MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte
}

func (s *Staker) GenerateProof(tb testing.TB, hash []byte) string {
	tb.Helper()
	proof, err := vrf.NewK256VRF().Prove(s.PrivKey, hash)
	require.NoError(tb, err)
	return hex.EncodeToString(proof)
}

func (f *Fixture) AddStakers(tb testing.TB, num int) []Staker {
	tb.Helper()

	stakers := make([]Staker, num)
	for i := 0; i < num; i++ {
		privKey := secp256k1.GenPrivKey()
		stakers[i] = Staker{
			PrivKey: privKey.Bytes(),
			PubKey:  hex.EncodeToString(privKey.PubKey().Bytes()),
			Address: privKey.PubKey().Address().Bytes(),
		}

		f.mintCoinsForAccount(tb, stakers[i].Address, 100)
		f.initAccountWithCoins(tb, stakers[i].Address, sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1e18))))

		// Add to allowlist.
		f.executeMsg(&types.MsgAddToAllowlist{
			Sender:    f.Creator.Address(),
			PublicKey: stakers[i].PubKey,
		}, f.Creator.addr, 0)

		// Stake.
		stakeMsg := &types.MsgStake{
			Sender:    stakers[i].Address.String(),
			PublicKey: stakers[i].PubKey,
			Memo:      "YWRkcmVzcw==",
			Stake:     sdk.NewCoin(BondDenom, math.NewIntFromUint64(1000000000000000000)),
		}
		stakeMsg.Proof = stakers[i].GenerateProof(f.tb, stakeMsg.MsgHash("", f.ChainID, 0))
		f.executeMsg(stakeMsg, stakers[i].Address, 0)

		// Second stake to test sequence number.
		stakeMsg2 := &types.MsgStake{
			Sender:    stakers[i].Address.String(),
			PublicKey: stakers[i].PubKey,
			Memo:      "YWRkcmVzcw==",
			Stake:     sdk.NewCoin(BondDenom, math.NewIntFromUint64(500000000000000000)),
		}
		stakeMsg2.Proof = stakers[i].GenerateProof(f.tb, stakeMsg.MsgHash("", f.ChainID, 1))
		f.executeMsg(stakeMsg2, stakers[i].Address, 0)
	}

	f.Stakers = append(f.Stakers, stakers...)
	return stakers
}

// AddStakers generates stakers and adds them to the allowlist. The
// stakers subsequently send their stakes to the core contract.
func (f *Fixture) AddStakersContract(tb testing.TB, num int) []Staker {
	tb.Helper()

	stakers := make([]Staker, num)
	for i := 0; i < num; i++ {
		privKey := secp256k1.GenPrivKey()
		stakers[i] = Staker{
			PrivKey: privKey.Bytes(),
			PubKey:  hex.EncodeToString(privKey.PubKey().Bytes()),
			Address: privKey.PubKey().Address().Bytes(),
		}

		f.mintCoinsForAccount(tb, stakers[i].Address, 100)

		// Add to allowlist.
		f.executeCoreContract(
			f.Creator.Address(),
			testutil.AddToAllowListMsgContract(stakers[i].PubKey),
			sdk.NewCoins(),
		)

		// Stake.
		f.initAccountWithCoins(tb, stakers[i].Address, sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1e18))))
		proof := f.generateStakeProof(tb, stakers[i].PrivKey, "YWRkcmVzcw==", 0)
		f.executeCoreContract(
			stakers[i].Address.String(),
			testutil.StakeMsgContract(stakers[i].PubKey, proof, "YWRkcmVzcw=="),
			sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(1000000000000000000))),
		)

		// Second stake to test sequence number.
		proof = f.generateStakeProof(tb, stakers[i].PrivKey, "YWRkcmVzcw==", 1)
		f.executeCoreContract(
			f.Creator.Address(),
			testutil.StakeMsgContract(stakers[i].PubKey, proof, "YWRkcmVzcw=="),
			sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(500000000000000000))),
		)
	}

	f.Stakers = append(f.Stakers, stakers...)
	return stakers
}

func (f *Fixture) AddDataProxy(tb testing.TB, proxyPubKey, payoutAddr string, proxyFee sdk.Coin) {
	tb.Helper()

	err := f.SetDataProxyConfig(proxyPubKey, payoutAddr, proxyFee)
	require.NoError(tb, err)
}

func (f *Fixture) DrainDataRequestPool(targetHeight uint64) []byte {
	return f.executeCoreContract(
		f.Creator.Address(),
		testutil.DrainDataRequestPoolMsgContract(targetHeight),
		sdk.NewCoins(sdk.NewCoin(BondDenom, math.NewIntFromUint64(500000000000000000))),
	)
}

// CheckDataRequestsByStatus checks that the given number of data requests is
// retrieved by GetDataRequestsByStatus.
func (f *Fixture) CheckDataRequestsByStatus(tb testing.TB, status types.DataRequestStatus, expectedTotal, fetchLimit uint32) {
	tb.Helper()

	remainingTotal := expectedTotal

	drs, lastSeen, total, err := f.CoreKeeper.GetDataRequestsByStatus(f.Context(), status, fetchLimit, nil)
	require.NoError(tb, err)
	require.Equal(tb, expectedTotal, total)
	require.Equal(tb, min(remainingTotal, fetchLimit), uint32(len(drs))) //nolint:gosec // G115: All positive integers

	for fetchLimit < remainingTotal {
		remainingTotal -= fetchLimit

		require.NotNil(tb, lastSeen)
		drs, lastSeen, total, err = f.CoreKeeper.GetDataRequestsByStatus(f.Context(), status, fetchLimit, lastSeen)
		require.NoError(tb, err)
		require.Equal(tb, expectedTotal, total)
		require.Equal(tb, min(remainingTotal, fetchLimit), uint32(len(drs))) //nolint:gosec // G115: All positive integers
	}
}

// generateStakeProof generates a proof for a stake message given a
// base64-encoded memo.
func (f *Fixture) generateStakeProof(tb testing.TB, signKey []byte, base64Memo string, seqNum uint64) string {
	tb.Helper()

	// We use legacy type to generate legacy hash.
	msg := types.MsgLegacyStake{
		Memo: base64Memo,
	}
	proof, err := vrf.NewK256VRF().Prove(signKey, msg.MsgHash(f.CoreContractAddr.String(), f.ChainID, seqNum))
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
	allBytes = append(allBytes, []byte(f.CoreContractAddr.String())...)

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

func (f *Fixture) createRevealMsg(staker Staker, revealBody types.RevealBody) (types.MsgReveal, string, string) {
	f.tb.Helper()

	revealBodyHash := revealBody.RevealBodyHash()

	proof := f.generateRevealProof(f.tb, staker.PrivKey, revealBodyHash)

	msg := types.MsgReveal{
		Sender:     staker.Address.String(),
		RevealBody: &revealBody,
		PublicKey:  staker.PubKey,
		Proof:      proof,
		Stderr:     []string{""},
		Stdout:     []string{""},
	}

	msg.Proof = staker.GenerateProof(f.tb, msg.MsgHash("", f.ChainID))
	return msg, hex.EncodeToString(msg.RevealHash()), proof
}

// createRevealMsg constructs and returns a reveal message and its corresponding
// commitment and proof.
func (f *Fixture) createRevealMsgContract(staker Staker, revealBody types.RevealBody) ([]byte, string, string) {
	f.tb.Helper()

	revealBodyHash := revealBody.RevealBodyHash()

	proof := f.generateRevealProof(f.tb, staker.PrivKey, revealBodyHash)

	msg := testutil.RevealMsgContract(
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
	allBytes = append(allBytes, []byte(f.CoreContractAddr.String())...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(tb, err)

	return hex.EncodeToString(proof)
}

// executeCommitOrReveal executes a commit msg or a reveal msg.
func (f *Fixture) executeCommitOrRevealContract(sender sdk.AccAddress, msg []byte, gasLimit uint64) {
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
