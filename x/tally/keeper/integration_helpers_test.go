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

// commitRevealDataRequest performs the following steps to prepare a
// tally-ready data request.
// 1. Generate staker key and add to allowlist.
// 2. Upload data request and tally oracle programs.
// 3. Create an account and stake.
// 4. Post a data request.
// 5. The staker commits and reveals.
// It returns the data request ID.
func (f *fixture) commitRevealDataRequest(t *testing.T, requestMemo string) string {
	// 1. Generate staker key and add to allowlist.
	privKey := secp256k1.GenPrivKey()
	stakerKey := privKey.Bytes()
	stakerPubKey := hex.EncodeToString(privKey.PubKey().Bytes())
	staker := privKey.PubKey().Address().Bytes()

	_, err := f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		f.deployer,
		[]byte(fmt.Sprintf(addToAllowListMsg, stakerPubKey)),
		sdk.NewCoins(),
	)
	require.NoError(t, err)

	// 2. Upload data request and tally oracle programs.
	execProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
	require.NoError(t, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testdata.SampleTallyWasm2(), f.Context().BlockTime(), f.Context().BlockHeight(), 1000)
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	// 3. Create an account and stake.
	f.initAccountWithCoins(t, staker, sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1e18))))

	proof := f.generateStakeProof(t, stakerKey)
	_, err = f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		staker,
		[]byte(fmt.Sprintf(stakeMsg, stakerPubKey, proof)),
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
	)
	require.NoError(t, err)

	// 4. Post a data request.
	resJSON, err := f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		f.deployer,
		[]byte(fmt.Sprintf(postDataRequestMsg, hex.EncodeToString(execProgram.Hash), hex.EncodeToString(tallyProgram.Hash), requestMemo)),
		sdk.NewCoins(),
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

	// 5. The staker commits and reveals.
	revealBody := types.RevealBody{
		ID:           drID,
		Salt:         []byte("9c0257114eb9399a2985f8e75dad7600c5d89fe3824ffa99ec1c3eb8bf3b0501"),
		Reveal:       "Ghkvq84TmIuEmU1ClubNxBjVXi8df5QhiNQEC5T8V6w=",
		GasUsed:      0,
		ExitCode:     0,
		ProxyPubKeys: []string{},
	}
	commitment, err := revealBody.TryHash()
	require.NoError(t, err)

	proof = f.generateCommitProof(t, stakerKey, drID, commitment, res.Height)
	_, err = f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		staker,
		[]byte(fmt.Sprintf(commitMsg, drID, commitment, stakerPubKey, proof)),
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
	)
	require.NoError(t, err)

	proof = f.generateRevealProof(t, stakerKey, drID, commitment, res.Height)
	_, err = f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		staker,
		[]byte(fmt.Sprintf(revealMsg, drID, drID, stakerPubKey, proof)),
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
	)
	require.NoError(t, err)

	return res.DrID
}

var addToAllowListMsg = `{
	"add_to_allowlist": {
	  "public_key": "%s"
    }
}`

var stakeMsg = `{
  "stake": {
    "public_key": "%s",
    "proof": "%s",
    "memo": "YWRkcmVzcw=="
  }
}`

var postDataRequestMsg = `{
	"post_data_request": {
	  "posted_dr": {
		"version": "0.0.1",
		"exec_program_id": "%s",
		"exec_inputs": "ZXhlY19pbnB1dHM=",
		"exec_gas_limit": 10,
		"tally_program_id": "%s",
		"tally_inputs": "dGFsbHlfaW5wdXRz",
		"tally_gas_limit": 10,
		"replication_factor": 1,
		"consensus_filter": "AA==",
		"gas_price": "10",
		"memo": "%s"
	  },
	  "seda_payload": "",
	  "payback_address": "AQID"
	}
}`

var commitMsg = `{
  "commit_data_result": {
    "dr_id": "%s",
    "commitment": "%s",
    "public_key": "%s",
    "proof": "%s"
  }
}`

var revealMsg = `{
  "reveal_data_result": {
    "dr_id": "%s",
    "reveal_body": {
      "id": "%s",
      "salt": "9c0257114eb9399a2985f8e75dad7600c5d89fe3824ffa99ec1c3eb8bf3b0501",
      "exit_code": 0,
      "gas_used": 0,
      "reveal": "Ghkvq84TmIuEmU1ClubNxBjVXi8df5QhiNQEC5T8V6w=",
      "proxy_public_keys": []
    },
    "public_key": "%s",
    "proof": "%s",
    "stderr": [],
    "stdout": []
  }
}`

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
