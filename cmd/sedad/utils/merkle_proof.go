package utils

import (
	"bytes"
	"fmt"
	"hash"

	"golang.org/x/crypto/sha3"
)

// GenerateProof returns the proof of inclusion for the entry at
// index i in the entries list.
func GenerateProof(entries [][]byte, i int) ([][]byte, error) {
	if i < 0 || i >= len(entries) {
		return nil, fmt.Errorf("index %d is out of bounds", i)
	}
	if len(entries) == 1 {
		return [][]byte{}, nil
	}

	k := largestPowerOfTwoSmallerThan(len(entries))
	recurse := entries[:k]
	aggregate := entries[k:]
	if i >= k {
		i -= k
		recurse, aggregate = aggregate, recurse
	}
	res, err := GenerateProof(recurse, i)
	if err != nil {
		return nil, err
	}
	res = append(res, RootFromEntries(aggregate))
	return res, nil
}

func VerifyProof(proof [][]byte, root, entry []byte) bool {
	return bytes.Equal(processProof(sha3.NewLegacyKeccak256(), proof, entry), root)
}

func processProof(sha hash.Hash, proof [][]byte, entry []byte) []byte {
	currentHash := leafHash(sha, entry)

	for i := 0; i < len(proof); i++ {
		if bytes.Compare(currentHash, proof[i]) == -1 {
			currentHash = nodeHash(sha, currentHash, proof[i])
		} else {
			currentHash = nodeHash(sha, proof[i], currentHash)
		}
	}
	return currentHash
}
