package types

import (
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// TODO Remove contractAddr
func (m MsgStake) ComputeStakeHash(contractAddr, chainID string, sequenceNum uint64) ([]byte, error) {
	memoBytes, err := hex.DecodeString(m.Memo)
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
	allBytes = append(allBytes, []byte(contractAddr)...)
	allBytes = append(allBytes, seqBytes...)

	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(allBytes)
	return hasher.Sum(nil), nil
}
