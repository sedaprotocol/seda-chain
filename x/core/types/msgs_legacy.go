package types

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/sha3"
)

// MsgHash computes the hash of a stake message for the old contract format
// that included the core contract address in the hash. The new format omits
// this field.
func (m MsgLegacyStake) MsgHash(coreContractAddr, chainID string, sequenceNum uint64) ([]byte, error) {
	memoBytes, err := base64.StdEncoding.DecodeString(m.Memo)
	if err != nil {
		return nil, err
	}
	memoHasher := sha3.NewLegacyKeccak256()
	memoHasher.Write(memoBytes)
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
	return hasher.Sum(nil), nil
}

func (m MsgLegacyUnstake) MsgHash(coreContractAddr, chainID string, sequenceNum uint64) ([]byte, error) {
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
	return hasher.Sum(nil), nil
}

func (m MsgLegacyWithdraw) MsgHash(coreContractAddr, chainID string, sequenceNum uint64) ([]byte, error) {
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
	return hasher.Sum(nil), nil
}

// MsgHash computes the hash of a commit message for the old contract format
// that included the core contract address in the hash. The new format omits
// this field.
func (m MsgLegacyCommit) MsgHash(contractAddr, chainID string, drHeight int64) ([]byte, error) {
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
	return hasher.Sum(nil), nil
}

func (m MsgLegacyReveal) Validate(config DataRequestConfig, replicationFactor uint32) error {
	// Ensure that the exit code fits within 8 bits (unsigned).
	if m.RevealBody.ExitCode > uint32(^uint8(0)) {
		return ErrInvalidRevealExitCode
	}

	revealSizeLimit := config.DrRevealSizeLimitInBytes / replicationFactor
	if len(m.RevealBody.Reveal) > int(revealSizeLimit) {
		return ErrRevealTooBig.Wrapf("%d bytes > %d bytes", len(m.RevealBody.Reveal), revealSizeLimit)
	}

	for _, key := range m.RevealBody.ProxyPubKeys {
		_, err := hex.DecodeString(key)
		if err != nil {
			return ErrInvalidProxyPublicKey.Wrapf("%s", err.Error())
		}
	}
	return nil
}

// RevealHash computes a hash of reveal contents to be used as a commitment
// by the executors.
func (m MsgLegacyReveal) RevealHash() ([]byte, error) {
	revealBodyHash, err := m.RevealBody.RevealBodyHash()
	if err != nil {
		return nil, err
	}

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte("reveal_message"))
	hasher.Write(revealBodyHash)
	hasher.Write([]byte(m.PublicKey))
	hasher.Write([]byte(m.Proof))
	hasher.Write([]byte(strings.Join(m.Stderr, "")))
	hasher.Write([]byte(strings.Join(m.Stdout, "")))

	return hasher.Sum(nil), nil
}

// MsgHash computes the hash of a reveal message for the old contract format
// that included the core contract address in the hash. The new format omits
// this field.
func (m MsgLegacyReveal) MsgHash(contractAddr, chainID string) ([]byte, error) {
	revealBodyHash, err := m.RevealBody.RevealBodyHash()
	if err != nil {
		return nil, err
	}

	allBytes := append([]byte("reveal_data_result"), revealBodyHash...)
	allBytes = append(allBytes, []byte(chainID)...)
	allBytes = append(allBytes, []byte(contractAddr)...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
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
