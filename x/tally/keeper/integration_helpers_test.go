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
	defaultRevealTimeoutBlocks = 10
)

type PostDataRequestResponse struct {
	DrID   string `json:"dr_id"`
	Height uint64 `json:"height"`
}

type commitRevealConfig struct {
	requestHeight uint64
	requestMemo   string
	reveal        string
	proxyPubKeys  []string
	gasUsed       uint64
	exitCode      byte
}

// commitRevealDataRequest simulates stakers committing and revealing
// for a data request. It returns the data request ID.
func (f *fixture) commitRevealDataRequest(t *testing.T, replicationFactor, numCommits, numReveals int, timeout bool, config commitRevealConfig) (string, []staker) {
	stakers := f.addStakers(t, 5)

	// Upload data request and tally oracle programs.
	execProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm(), f.Context().BlockTime())
	err := f.wasmStorageKeeper.OracleProgram.Set(f.Context(), execProgram.Hash, execProgram)
	require.NoError(t, err)

	tallyProgram := wasmstoragetypes.NewOracleProgram(testwasms.SampleTallyWasm2(), f.Context().BlockTime())
	err = f.wasmStorageKeeper.OracleProgram.Set(f.Context(), tallyProgram.Hash, tallyProgram)
	require.NoError(t, err)

	// Post a data request.
	res, err := f.postDataRequest(execProgram.Hash, tallyProgram.Hash, config.requestMemo, replicationFactor)
	require.NoError(t, err)

	drID := res.DrID

	// The stakers commit and reveal.
	revealMsgs, err := f.commitDataRequest(stakers, res.Height, drID, numCommits, config)
	require.NoError(t, err)

	err = f.revealDataRequest(stakers, revealMsgs[:numReveals])
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

// commitDataRequest executes a commit for each of the given stakers and returns
// a list of corresponding reveal messages.
func (f *fixture) commitDataRequest(stakers []staker, height uint64, drID string, numCommits int, config commitRevealConfig) ([][]byte, error) {
	revealBody := types.RevealBody{
		RequestID:    drID,
		Reveal:       config.reveal,
		GasUsed:      config.gasUsed,
		ExitCode:     config.exitCode,
		ProxyPubKeys: config.proxyPubKeys,
	}

	var revealMsgs [][]byte
	for i := 0; i < numCommits; i++ {
		revealMsg, commitment, _, err := f.createRevealMsg(stakers[i], revealBody)
		if err != nil {
			return nil, err
		}

		proof, err := f.generateCommitProof(stakers[i].key, drID, commitment, height)
		if err != nil {
			return nil, err
		}
		_, err = f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			commitMsg(drID, commitment, stakers[i].pubKey, proof, config.gasUsed),
			sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntFromUint64(1))),
		)
		if err != nil {
			return nil, err
		}

		revealMsgs = append(revealMsgs, revealMsg)
	}

	return revealMsgs, nil
}

func (f *fixture) revealDataRequest(stakers []staker, revealMsgs [][]byte) error {
	for i := 0; i < len(revealMsgs); i++ {
		_, err := f.contractKeeper.Execute(
			f.Context(),
			f.coreContractAddr,
			stakers[i].address,
			revealMsgs[i],
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

func revealMsg(drID, reveal, stakerPubKey, proof string, proxyPubKeys []string, exitCode byte, drHeight, gasUsed uint64) []byte {
	quotedObjects := []string{}
	for _, obj := range proxyPubKeys {
		quotedObjects = append(quotedObjects, fmt.Sprintf("%q", obj))
	}
	pks := strings.Join(quotedObjects, ",")

	return []byte(fmt.Sprintf(`{
		"reveal_data_result": {
		  "reveal_body": {
			"dr_id": "%s",
			"dr_block_height": %d,
			"exit_code": %d,
			"gas_used": %d,
			"reveal": "%s",
			"proxy_public_keys": [%s]
		  },
		  "public_key": "%s",
		  "proof": "%s",
		  "stderr": [],
		  "stdout": []
		}
	}`, drID, drHeight, exitCode, gasUsed, reveal, pks, stakerPubKey, proof))
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

func (f *fixture) initAccountWithCoins(t *testing.T, addr sdk.AccAddress, coins sdk.Coins) {
	err := f.bankKeeper.MintCoins(f.Context(), minttypes.ModuleName, coins)
	require.NoError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.Context(), minttypes.ModuleName, addr, coins)
	require.NoError(t, err)
}

// generateRevealBodyHash generates the hash of a given reveal body.
// Since the RevealBody type in the tally module does not include the
// salt field, the salt must be provided separately.
func (f *fixture) generateRevealBodyHash(rb types.RevealBody) ([]byte, error) {
	revealHasher := sha3.NewLegacyKeccak256()
	revealBytes, err := base64.StdEncoding.DecodeString(rb.Reveal)
	if err != nil {
		return nil, err
	}
	revealHasher.Write(revealBytes)
	revealHash := revealHasher.Sum(nil)

	hasher := sha3.NewLegacyKeccak256()

	idBytes, err := hex.DecodeString(rb.RequestID)
	if err != nil {
		return nil, err
	}
	hasher.Write(idBytes)

	reqHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(reqHeightBytes, rb.RequestBlockHeight)
	hasher.Write(reqHeightBytes)

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

	return hasher.Sum(nil), nil
}

// createRevealMsg constructs and returns a reveal message and its corresponding
// commitment and proof.
func (f *fixture) createRevealMsg(staker staker, revealBody types.RevealBody) ([]byte, string, string, error) {
	revealBodyHash, err := f.generateRevealBodyHash(revealBody)
	if err != nil {
		return nil, "", "", err
	}
	proof, err := generateRevealProof(staker.key, revealBodyHash, f.chainID, f.coreContractAddr.String())
	if err != nil {
		return nil, "", "", err
	}

	msg := revealMsg(
		revealBody.RequestID,
		revealBody.Reveal,
		staker.pubKey,
		proof,
		revealBody.ProxyPubKeys,
		revealBody.ExitCode,
		revealBody.RequestBlockHeight,
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

	return msg, hex.EncodeToString(commitment), proof, nil
}

func generateRevealProof(signKey []byte, revealBodyHash []byte, chainID, coreContractAddr string) (string, error) {
	revealBytes := []byte("reveal_data_result")

	allBytes := append(revealBytes, revealBodyHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(coreContractAddr)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	hash := hasher.Sum(nil)

	proof, err := vrf.NewK256VRF().Prove(signKey, hash)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(proof), nil
}
