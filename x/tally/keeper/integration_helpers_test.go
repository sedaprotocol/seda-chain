package keeper_test

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/sedaprotocol/seda-chain/x/tally/keeper/testdata"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	salt                       = "9c0257114eb9399a2985f8e75dad7600c5d89fe3824ffa99ec1c3eb8bf3b0501"
	defaultRevealTimeoutBlocks = 10
)

// commitRevealDataRequest simulates stakers committing and revealing
// for a data request. It returns the data request ID.
func (f *fixture) commitRevealDataRequest(t *testing.T, requestMemo, reveal string, replicationFactor, numCommits, numReveals int, timeout bool) string {
	stakers := f.addStakers(t, 5)

	// Upload data request and tally oracle programs.
	execProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	err := f.wasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
	require.NoError(t, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm2(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	// Post a data request.
	resJSON, err := f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		stakers[0].address,
		postDataRequestMsg(execProgram.Hash, tallyProgram.Hash, requestMemo, replicationFactor),
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(3000000000000100))),
	)
	require.NoError(t, err)

	type PostDataRequestResponse struct {
		DrID   string `json:"dr_id"`
		Height uint64 `json:"height"`
	}
	var res PostDataRequestResponse
	err = json.Unmarshal(resJSON, &res)
	require.NoError(t, err)
	drID := res.DrID

	// The stakers commit and reveal.
	revealBody := types.RevealBody{
		ID:           drID,
		Salt:         []byte(salt),
		Reveal:       reveal,
		GasUsed:      0,
		ExitCode:     0,
		ProxyPubKeys: []string{},
	}
	commitment, err := revealBody.TryHash()
	require.NoError(t, err)

	for i := 0; i < numCommits; i++ {
		proof := f.generateCommitProof(t, stakers[i].key, drID, commitment, res.Height)
		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			commitMsg(drID, commitment, stakers[i].pubKey, proof),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		require.NoError(t, err)
	}

	for i := 0; i < numReveals; i++ {
		proof := f.generateRevealProof(t, stakers[i].key, drID, commitment, res.Height)
		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			revealMsg(drID, reveal, stakers[i].pubKey, proof),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		require.NoError(t, err)
	}

	if timeout {
		for i := 0; i < defaultRevealTimeoutBlocks; i++ {
			f.AddBlock()
		}
	}

	return res.DrID
}

type staker struct {
	key     []byte
	pubKey  string
	address []byte
}

// addStakers generates stakers and adds them to the allowlist. The
// stakers subsequently send their stakes to the core contract.
func (f *fixture) addStakers(t *testing.T, num int) []staker {
	stakers := make([]staker, num)
	for i := 0; i < num; i++ {
		privKey := secp256k1.GenPrivKey()
		stakers[i] = staker{
			key:     privKey.Bytes(),
			pubKey:  hex.EncodeToString(privKey.PubKey().Bytes()),
			address: privKey.PubKey().Address().Bytes(),
		}

		_, err := f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			f.deployer,
			addToAllowListMsg(stakers[i].pubKey),
			sdk.NewCoins(),
		)
		require.NoError(t, err)

		f.initAccountWithCoins(t, stakers[i].address, sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1e18))))

		proof := f.generateStakeProof(t, stakers[i].key)
		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			stakeMsg(stakers[i].pubKey, proof),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		require.NoError(t, err)
	}
	return stakers
}

func addToAllowListMsg(stakerPubKey string) []byte {
	var addToAllowListMsg = `{
		"add_to_allowlist": {
		  "public_key": "%s"
		}
	}`
	return []byte(fmt.Sprintf(addToAllowListMsg, stakerPubKey))
}

func stakeMsg(stakerPubKey, proof string) []byte {
	var stakeMsg = `{
		"stake": {
		  "public_key": "%s",
		  "proof": "%s",
		  "memo": "YWRkcmVzcw=="
		}
	}`
	return []byte(fmt.Sprintf(stakeMsg, stakerPubKey, proof))
}

func postDataRequestMsg(execProgHash, tallyProgHash []byte, requestMemo string, replicationFactor int) []byte {
	var postDataRequestMsg = `{
		"post_data_request": {
		  "posted_dr": {
			"version": "0.0.1",
			"exec_program_id": "%s",
			"exec_inputs": "ZXhlY19pbnB1dHM=",
			"exec_gas_limit": 10,
			"tally_program_id": "%s",
			"tally_inputs": "dGFsbHlfaW5wdXRz",
			"tally_gas_limit": 300000000000000,
			"replication_factor": %d,
			"consensus_filter": "AA==",
			"gas_price": "10",
			"memo": "%s"
		  },
		  "seda_payload": "",
		  "payback_address": "AQID"
		}
	}`
	return []byte(fmt.Sprintf(postDataRequestMsg, hex.EncodeToString(execProgHash), hex.EncodeToString(tallyProgHash), replicationFactor, requestMemo))
}

func commitMsg(drID, commitment, stakerPubKey, proof string) []byte {
	var commitMsg = `{
		"commit_data_result": {
		  "dr_id": "%s",
		  "commitment": "%s",
		  "public_key": "%s",
		  "proof": "%s"
		}
	}`
	return []byte(fmt.Sprintf(commitMsg, drID, commitment, stakerPubKey, proof))
}

func revealMsg(drID, reveal, stakerPubKey, proof string) []byte {
	var revealMsg = `{
		"reveal_data_result": {
		  "dr_id": "%s",
		  "reveal_body": {
			"id": "%s",
			"salt": "%s",
			"exit_code": 0,
			"gas_used": 0,
			"reveal": "%s",
			"proxy_public_keys": []
		  },
		  "public_key": "%s",
		  "proof": "%s",
		  "stderr": [],
		  "stdout": []
		}
	}`
	return []byte(fmt.Sprintf(revealMsg, drID, drID, salt, reveal, stakerPubKey, proof))
}

// generateStakeProof generates a proof for a stake message given a
// base64-encoded memo.
func (f *fixture) generateStakeProof(t *testing.T, signKey []byte) string {
	// TODO
	// var sequence uint64 = 0

	memo := "YWRkcmVzcw=="
	memoBytes, err := base64.StdEncoding.DecodeString(memo)
	require.NoError(t, err)

	// Create slices for each component
	stakeBytes := []byte("stake")

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(memoBytes)
	memoHash := hasher.Sum(nil)

	chainIDBytes := []byte(f.chainID)
	contractAddrBytes := []byte(f.coreContractAddr.String())

	sequenceBytes := make([]byte, 16)
	// binary.BigEndian.PutUint64(sequenceBytes, sequence) // TODO

	allBytes := append([]byte{}, stakeBytes...)
	allBytes = append(allBytes, memoHash...)
	allBytes = append(allBytes, chainIDBytes...)
	allBytes = append(allBytes, contractAddrBytes...)
	allBytes = append(allBytes, sequenceBytes...)

	hasher.Reset()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(t, err)
	return hex.EncodeToString(proof)
}

func (f *fixture) generateCommitProof(t *testing.T, signKey []byte, drID, commitment string, drHeight uint64) string {
	commitBytes := []byte("commit_data_result")
	drIDBytes := []byte(drID)

	drHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(drHeightBytes, drHeight)

	commitmentBytes := []byte(commitment)
	chainIDBytes := []byte(f.chainID)
	contractAddrBytes := []byte(f.coreContractAddr.String())

	allBytes := append([]byte{}, commitBytes...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, commitmentBytes...)
	allBytes = append(allBytes, chainIDBytes...)
	allBytes = append(allBytes, contractAddrBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(t, err)

	return hex.EncodeToString(proof)
}

func (f *fixture) generateRevealProof(t *testing.T, signKey []byte, drID, revealBodyHash string, drHeight uint64) string {
	revealBytes := []byte("reveal_data_result")
	drIDBytes := []byte(drID)

	drHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(drHeightBytes, drHeight)

	revealBodyHashBytes, err := hex.DecodeString(revealBodyHash)
	require.NoError(t, err)

	chainIDBytes := []byte(f.chainID)
	contractAddrBytes := []byte(f.coreContractAddr.String())

	allBytes := append([]byte{}, revealBytes...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, revealBodyHashBytes...)
	allBytes = append(allBytes, chainIDBytes...)
	allBytes = append(allBytes, contractAddrBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	require.NoError(t, err)
	return hex.EncodeToString(proof)
}

func (f *fixture) initAccountWithCoins(t *testing.T, addr sdk.AccAddress, coins sdk.Coins) {
	err := f.bankKeeper.MintCoins(f.Context(), minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, addr, coins)
	require.NoError(t, err)
}
