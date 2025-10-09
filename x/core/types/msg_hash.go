package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/sha3"
)

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func mustDecodeBase64(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

// ComputeDataRequestID computes the hex-encoded hash of the PostDataRequest
// message to be used as the data request ID.
func (m *MsgPostDataRequest) ComputeDataRequestID() (string, error) {
	execProgramIDBytes, err := hex.DecodeString(m.ExecProgramID)
	if err != nil {
		return "", err
	}
	tallyProgramIDBytes, err := hex.DecodeString(m.TallyProgramID)
	if err != nil {
		return "", err
	}

	execInputsHasher := sha3.NewLegacyKeccak256()
	execInputsHasher.Write(m.ExecInputs)
	execInputsHash := execInputsHasher.Sum(nil)

	tallyInputsHasher := sha3.NewLegacyKeccak256()
	tallyInputsHasher.Write(m.TallyInputs)
	tallyInputsHash := tallyInputsHasher.Sum(nil)

	consensusFilterHasher := sha3.NewLegacyKeccak256()
	consensusFilterHasher.Write(m.ConsensusFilter)
	consensusFilterHash := consensusFilterHasher.Sum(nil)

	memoHasher := sha3.NewLegacyKeccak256()
	memoHasher.Write(m.Memo)
	memoHash := memoHasher.Sum(nil)

	execGasLimitBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(execGasLimitBytes, m.ExecGasLimit)
	tallyGasLimitBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tallyGasLimitBytes, m.TallyGasLimit)
	replicationFactorBytes := make([]byte, 2)
	//nolint:gosec // G115: Replication factor is guaranteed to fit in uint16.
	binary.BigEndian.PutUint16(replicationFactorBytes, uint16(m.ReplicationFactor))

	dataRequestHasher := sha3.NewLegacyKeccak256()
	dataRequestHasher.Write([]byte(m.Version))
	dataRequestHasher.Write(execProgramIDBytes)
	dataRequestHasher.Write(execInputsHash)
	dataRequestHasher.Write(execGasLimitBytes)
	dataRequestHasher.Write(tallyProgramIDBytes)
	dataRequestHasher.Write(tallyInputsHash)
	dataRequestHasher.Write(tallyGasLimitBytes)
	dataRequestHasher.Write(replicationFactorBytes)
	dataRequestHasher.Write(consensusFilterHash)
	dataRequestHasher.Write(m.GasPrice.BigInt().Bytes())
	dataRequestHasher.Write(memoHash)

	return hex.EncodeToString(dataRequestHasher.Sum(nil)), nil
}

func (m MsgStake) MsgHash(_, chainID string, sequenceNum uint64) []byte {
	memoHasher := sha3.NewLegacyKeccak256()
	// MsgStake.Validate ensures Memo is valid base64.
	memoHasher.Write(mustDecodeBase64(m.Memo))
	memoHash := memoHasher.Sum(nil)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)

	allBytes := append([]byte{}, []byte("stake")...)
	allBytes = append(allBytes, memoHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

// MsgHash computes the hash of a stake message for the old contract format
// that included the core contract address in the hash. The new format omits
// this field.
func (m MsgLegacyStake) MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte {
	memoHasher := sha3.NewLegacyKeccak256()
	// MsgLegacyStake.Validate ensures Memo is valid base64.
	memoHasher.Write(mustDecodeBase64(m.Memo))
	memoHash := memoHasher.Sum(nil)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)

	allBytes := append([]byte{}, []byte("stake")...)
	allBytes = append(allBytes, memoHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(coreContractAddr)...)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (m MsgUnstake) MsgHash(_, chainID string, sequenceNum uint64) []byte {
	allBytes := append([]byte{}, []byte("unstake")...)
	allBytes = append(allBytes, []byte(chainID)...)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (m MsgLegacyUnstake) MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte {
	allBytes := append([]byte{}, []byte("unstake")...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(coreContractAddr)...)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (m MsgWithdraw) MsgHash(_, chainID string, sequenceNum uint64) []byte {
	allBytes := append([]byte{}, []byte("withdraw")...)
	allBytes = append(allBytes, []byte(m.WithdrawAddress)...)
	allBytes = append(allBytes, []byte(chainID)...)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (m MsgLegacyWithdraw) MsgHash(coreContractAddr, chainID string, sequenceNum uint64) []byte {
	allBytes := append([]byte{}, []byte("withdraw")...)
	allBytes = append(allBytes, []byte(m.WithdrawAddress)...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(coreContractAddr)...)

	// Write to last 8 bytes of 16-byte variable.
	// TODO contract used uint128
	seqBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(seqBytes[8:], sequenceNum)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (m MsgCommit) MsgHash(_, chainID string, drHeight int64) []byte {
	drHeightBytes := make([]byte, 8)
	//nolint:gosec // G115: Block height is never negative.
	binary.BigEndian.PutUint64(drHeightBytes, uint64(drHeight))

	allBytes := append([]byte{}, []byte("commit_data_result")...)
	allBytes = append(allBytes, []byte(m.DrID)...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, m.Commit...)
	allBytes = append(allBytes, []byte(chainID)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

// MsgHash computes the hash of a commit message for the old contract format
// that included the core contract address in the hash. The new format omits
// this field.
func (m MsgLegacyCommit) MsgHash(contractAddr, chainID string, drHeight int64) []byte {
	drHeightBytes := make([]byte, 8)
	//nolint:gosec // G115: Block height is never negative.
	binary.BigEndian.PutUint64(drHeightBytes, uint64(drHeight))

	allBytes := append([]byte{}, []byte("commit_data_result")...)
	allBytes = append(allBytes, []byte(m.DrID)...)
	allBytes = append(allBytes, drHeightBytes...)
	allBytes = append(allBytes, m.Commit...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(contractAddr)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (m MsgReveal) MsgHash(_, chainID string) []byte {
	allBytes := append([]byte("reveal_data_result"), m.RevealBody.RevealBodyHash()...)
	allBytes = append(allBytes, []byte(chainID)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

// MsgHash computes the hash of a reveal message for the old contract format
// that included the core contract address in the hash. The new format omits
// this field.
func (m MsgLegacyReveal) MsgHash(contractAddr, chainID string) []byte {
	allBytes := append([]byte("reveal_data_result"), m.RevealBody.RevealBodyHash()...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(contractAddr)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

// RevealHash computes a hash of reveal contents to be used as a commitment
// by the executors.
func (m MsgReveal) RevealHash() []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("reveal_message"))
	hasher.Write(m.RevealBody.RevealBodyHash())
	hasher.Write([]byte(m.PublicKey))
	hasher.Write([]byte(m.Proof))
	hasher.Write([]byte(strings.Join(m.Stderr, "")))
	hasher.Write([]byte(strings.Join(m.Stdout, "")))

	return hasher.Sum(nil)
}

// RevealHash computes a hash of reveal contents to be used as a commitment
// by the executors.
func (m MsgLegacyReveal) RevealHash() []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("reveal_message"))
	hasher.Write(m.RevealBody.RevealBodyHash())
	hasher.Write([]byte(m.PublicKey))
	hasher.Write([]byte(m.Proof))
	hasher.Write([]byte(strings.Join(m.Stderr, "")))
	hasher.Write([]byte(strings.Join(m.Stdout, "")))

	return hasher.Sum(nil)
}

func (rb RevealBody) RevealBodyHash() []byte {
	revealHasher := sha3.NewLegacyKeccak256()
	revealHasher.Write(rb.Reveal)
	revealHash := revealHasher.Sum(nil)

	hasher := sha3.NewLegacyKeccak256()

	// MsgReveal.Validate ensures DrID is valid hex.
	hasher.Write(mustDecodeHex(rb.DrID))

	reqHeightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(reqHeightBytes, rb.DrBlockHeight)
	hasher.Write(reqHeightBytes)

	// TODO RevealBody validator should bind rb.ExitCode value?
	hasher.Write([]byte{byte(rb.ExitCode)})

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

	return hasher.Sum(nil)
}

func (q QueryGetExecutorEligibilityRequest) MsgHash(chainID string) []byte {
	_, drIDBytes, _, _ := q.Parts()
	allBytes := append([]byte{}, []byte("is_executor_eligible")...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, []byte(chainID)...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}

func (q QueryGetExecutorEligibilityRequest) LegacyMsgHash(contractAddr string) []byte {
	_, drIDBytes, _, _ := q.Parts()
	allBytes := append([]byte{}, []byte("is_executor_eligible")...)
	allBytes = append(allBytes, drIDBytes...)
	allBytes = append(allBytes, []byte(contractAddr)...)
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil)
}
