package types

import (
	"encoding/base64"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

// ComputeLegacyStakeHash computes the hash of a stake message for the old contract
// format that included the core contract address in the hash. The new format omits
// this field.
func (m MsgStake) ComputeLegacyStakeHash(coreContractAddr, chainID string, sequenceNum uint64) ([]byte, error) {
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

func (m MsgStake) ComputeStakeHash(chainID string, sequenceNum uint64) ([]byte, error) {
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
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}
