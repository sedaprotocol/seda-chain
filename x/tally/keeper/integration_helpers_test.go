package keeper_test

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	vrf "github.com/sedaprotocol/vrf-go"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/sedaprotocol/seda-chain/testutil/testwasms"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
	wasmstoragetypes "github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	saltHex                    = "9c0257114eb9399a2985f8e75dad7600c5d89fe3824ffa99ec1c3eb8bf3b0501"
	defaultRevealTimeoutBlocks = 10
)

type PostDataRequestResponse struct {
	DrID   string `json:"dr_id"`
	Height uint64 `json:"height"`
}

// commitRevealDataRequest simulates stakers committing and revealing
// for a data request. It returns the data request ID.
func (f *fixture) commitRevealDataRequest(t *testing.T, requestMemo, reveal string, proxyPubKeys []string, gasUsed uint64, replicationFactor, numCommits, numReveals int, timeout bool) (string, []staker) {
	stakers := f.addStakers(t, 5)

	// Upload data request and tally oracle programs.
	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())
	err := f.wasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
	require.NoError(t, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm2(), f.Context().BlockTime())
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	// Post a data request.
	res, err := f.postDataRequest(execProgram.Hash, tallyProgram.Hash, requestMemo, replicationFactor)
	require.NoError(t, err)

	drID := res.DrID

	// The stakers commit and reveal.
	commitment, err := f.commitDataRequest(stakers, res.Height, gasUsed, drID, reveal, proxyPubKeys, numCommits)
	require.NoError(t, err)

	err = f.revealDataRequest(stakers, res.Height, gasUsed, drID, reveal, proxyPubKeys, commitment, numReveals)
	require.NoError(t, err)

	if timeout {
		for i := 0; i < defaultRevealTimeoutBlocks; i++ {
			f.AddBlock()
		}
	}
	return res.DrID, stakers
}

func (f *fixture) postDataRequest(execProgHash, tallyProgHash []byte, requestMemo string, replicationFactor int) (PostDataRequestResponse, error) {
	resJSON, err := f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		f.deployer,
		postDataRequestMsg(execProgHash, tallyProgHash, requestMemo, replicationFactor),
		sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1500000000000000000))),
	)
	if err != nil {
		return PostDataRequestResponse{}, err
	}

	var res PostDataRequestResponse
	err = json.Unmarshal(resJSON, &res)
	if err != nil {
		return PostDataRequestResponse{}, err
	}

	return res, nil
}

func (f *fixture) commitDataRequest(stakers []staker, height, gasUsed uint64, drID, reveal string, proxyPubKeys []string, numCommits int) (string, error) {
	revealBody := types.RevealBody{
		ID:           drID,
		Reveal:       reveal,
		GasUsed:      gasUsed,
		ExitCode:     0,
		ProxyPubKeys: proxyPubKeys,
	}
	commitment, err := f.generateRevealBodyHash(revealBody, saltHex)
	if err != nil {
		return "", err
	}

	for i := 0; i < numCommits; i++ {
		proof, err := f.generateCommitProof(stakers[i].key, drID, commitment, height)
		if err != nil {
			return "", err
		}

		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			commitMsg(drID, commitment, stakers[i].pubKey, proof, gasUsed),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		if err != nil {
			return "", err
		}
	}

	return commitment, nil
}

func (f *fixture) revealDataRequest(stakers []staker, height, gasUsed uint64, drID, reveal string, proxyPubKeys []string, commitment string, numReveals int) error {
	for i := 0; i < numReveals; i++ {
		proof, err := f.generateRevealProof(stakers[i].key, drID, commitment, height)
		if err != nil {
			return err
		}

		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			revealMsg(drID, reveal, stakers[i].pubKey, proof, proxyPubKeys, gasUsed),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		if err != nil {
			return err
		}
	}

	return nil
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

func (f *fixture) pauseContract(t *testing.T) {
	_, err := f.contractKeeper.Execute(
		f.Context(),
		f.coreContractAddr,
		f.deployer,
		[]byte(`{"pause":{}}`),
		sdk.NewCoins(),
	)
	require.NoError(t, err)
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
			"exec_gas_limit": 100000000000000000,
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

func commitMsg(drID, commitment, stakerPubKey, proof string, gasUsed uint64) []byte {
	var commitMsg = `{
		"commit_data_result": {
		  "dr_id": "%s",
		  "commitment": "%s",
		  "public_key": "%s",
		  "proof": "%s",
		  "gas_used": %d
		}
	}`
	return []byte(fmt.Sprintf(commitMsg, drID, commitment, stakerPubKey, proof, gasUsed))
}

func revealMsg(drID, reveal, stakerPubKey, proof string, proxyPubKeys []string, gasUsed uint64) []byte {
	quotedObjects := []string{}
	for _, obj := range proxyPubKeys {
		quotedObjects = append(quotedObjects, fmt.Sprintf("%q", obj))
	}
	pks := strings.Join(quotedObjects, ",")

	return []byte(fmt.Sprintf(`{
		"reveal_data_result": {
		  "dr_id": "%s",
		  "reveal_body": {
			"id": "%s",
			"salt": "%s",
			"exit_code": 0,
			"gas_used": %d,
			"reveal": "%s",
			"proxy_public_keys": [%s]
		  },
		  "public_key": "%s",
		  "proof": "%s",
		  "stderr": [],
		  "stdout": []
		}
	}`, drID, drID, saltHex, gasUsed, reveal, pks, stakerPubKey, proof))
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

func (f *fixture) generateCommitProof(signKey []byte, drID, commitment string, drHeight uint64) (string, error) {
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
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(proof), nil
}

func (f *fixture) generateRevealProof(signKey []byte, drID, revealBodyHash string, drHeight uint64) (string, error) {
	revealBytes := []byte("reveal_data_result")
	drIDBytes := []byte(drID)

	drHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(drHeightBytes, drHeight)

	revealBodyHashBytes, err := hex.DecodeString(revealBodyHash)
	if err != nil {
		return "", err
	}

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
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(proof), nil
}

func (f *fixture) initAccountWithCoins(t *testing.T, addr sdk.AccAddress, coins sdk.Coins) {
	err := f.bankKeeper.MintCoins(f.Context(), minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, addr, coins)
	require.NoError(t, err)
}

// generateRevealBodyHash generates the hash of a given reveal body.
// Since the RevealBody type in the tally module does not include the
// salt field, the salt must be provided separately.
func (f *fixture) generateRevealBodyHash(rb types.RevealBody, salt string) (string, error) {
	revealHasher := sha3.NewLegacyKeccak256()
	revealBytes, err := base64.StdEncoding.DecodeString(rb.Reveal)
	if err != nil {
		return "", err
	}
	revealHasher.Write(revealBytes)
	revealHash := revealHasher.Sum(nil)

	hasher := sha3.NewLegacyKeccak256()

	idBytes, err := hex.DecodeString(rb.ID)
	if err != nil {
		return "", err
	}
	hasher.Write(idBytes)

	hasher.Write([]byte(salt))
	hasher.Write([]byte{rb.ExitCode})

	gasUsedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(gasUsedBytes, rb.GasUsed)
	hasher.Write(gasUsedBytes)

	hasher.Write(revealHash)

	proxyPubKeyHasher := sha3.NewLegacyKeccak256()
	for _, key := range rb.ProxyPubKeys {
		keyHasher := sha3.NewLegacyKeccak256()
		keyHasher.Write([]byte(key))
		proxyPubKeyHasher.Write(keyHasher.Sum(nil))
	}
	hasher.Write(proxyPubKeyHasher.Sum(nil))

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
