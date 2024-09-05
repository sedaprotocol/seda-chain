// Largely taken from the CometBFT repo.
// Source: https://github.com/cometbft/cometbft/blob/main/crypto/merkle/tree.go
package utils

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"math/bits"

	"github.com/cometbft/cometbft/crypto/tmhash"
)

// TODO: make these have a large predefined capacity
var (
	leafPrefix  = []byte{0}
	innerPrefix = []byte{1}
)

// HashFromByteSlices computes a Merkle tree where the leaves are the byte slice,
// in the provided order. It follows RFC-6962.
func HashFromByteSlices(items [][]byte) []byte {
	return hashFromByteSlices(sha256.New(), items)
}

func hashFromByteSlices(sha hash.Hash, items [][]byte) []byte {
	switch len(items) {
	case 0:
		return emptyHash()
	case 1:
		return leafHashOpt(sha, items[0])
	default:
		k := getSplitPoint(int64(len(items)))
		a := hashFromByteSlices(sha, items[:k])
		b := hashFromByteSlices(sha, items[k:])

		var left, right []byte
		if bytes.Compare(a, b) == -1 {
			left, right = a, b
		} else {
			right, left = a, b
		}
		return innerHashOpt(sha, left, right)
	}
}

// returns tmhash(<empty>)
func emptyHash() []byte {
	return tmhash.Sum([]byte{})
}

// returns tmhash(0x00 || leaf)
func leafHashOpt(s hash.Hash, leaf []byte) []byte {
	s.Reset()
	s.Write(leafPrefix)
	s.Write(leaf)
	return s.Sum(nil)
}

func innerHashOpt(s hash.Hash, left []byte, right []byte) []byte {
	s.Reset()
	s.Write(innerPrefix)
	s.Write(left)
	s.Write(right)
	return s.Sum(nil)
}

// getSplitPoint returns the largest power of 2 less than length
func getSplitPoint(length int64) int64 {
	if length < 1 {
		panic("Trying to split a tree with size < 1")
	}
	uLength := uint(length)
	bitlen := bits.Len(uLength)
	k := int64(1 << uint(bitlen-1))
	if k == length {
		k >>= 1
	}
	return k
}
