package utils

import (
	"bytes"
	"hash"
	"math/bits"

	"golang.org/x/crypto/sha3"
)

// RootFromEntries computes the root of the Merkle Hash Tree whose
// leaves are the entries given in byte slices in the provided order.
// It largely follows RFC-6962 with some modifications:
//   - It uses SHA3-256 hashing algorithm.
//   - Leaves are double-hashed for second preimage resistance. No hashing
//     prefixes are used.
//   - Each hash pair is sorted to make proofs more succinct.
//   - "Super root" computation is supported.
func RootFromEntries(entries [][]byte) []byte {
	return rootFromEntries(sha3.New256(), entries)
}

func rootFromEntries(sha hash.Hash, entries [][]byte) []byte {
	switch len(entries) {
	case 0:
		return emptyHash(sha)
	case 1:
		return leafHash(sha, entries[0])
	default:
		k := largestPowerOfTwoSmallerThan(len(entries))
		a := rootFromEntries(sha, entries[:k])
		b := rootFromEntries(sha, entries[k:])

		if bytes.Compare(a, b) == -1 {
			return nodeHash(sha, a, b)
		}
		return nodeHash(sha, b, a)
	}
}

// SuperRoot computes the merkle parent of two existing merkle roots.
func SuperRoot(root1, root2 []byte) []byte {
	sha := sha3.New256()
	return parentHash(sha, root1, root2)
}

// SuperRootWithLeaf computes the merkle parent of an existing root
// and a new (unhashed) entry in a byte slice.
func SuperRootWithEntry(root, entry []byte) []byte {
	sha := sha3.New256()
	var hashedLeaf []byte
	if len(entry) == 0 {
		hashedLeaf = emptyHash(sha)
	} else {
		hashedLeaf = leafHash(sha, entry)
	}
	return parentHash(sha, hashedLeaf, root)
}

// parentHash computes the parent's hash given its two children a and b
// that have already been hashed.
func parentHash(sha hash.Hash, a, b []byte) []byte {
	if bytes.Compare(a, b) == -1 {
		return nodeHash(sha, a, b)
	}
	return nodeHash(sha, b, a)
}

// emptyHash returns keccak(<empty>)
func emptyHash(s hash.Hash) []byte {
	s.Reset()
	s.Write([]byte{})
	return s.Sum(nil)
}

// leafHash returns keccak(keccak(leaf))
func leafHash(s hash.Hash, leaf []byte) []byte {
	s.Reset()
	s.Write(leaf)
	single := s.Sum(nil)
	s.Reset()
	s.Write(single)
	return s.Sum(nil)
}

// nodeHash returns keccak(left || right)
func nodeHash(s hash.Hash, left []byte, right []byte) []byte {
	s.Reset()
	s.Write(left)
	s.Write(right)
	return s.Sum(nil)
}

// largestPowerOfTwoSmallerThan returns the largest power of two
// smaller than n.
func largestPowerOfTwoSmallerThan(n int) int {
	if n < 1 {
		panic("Trying to split a tree with size < 1")
	}
	uLength := uint(n)
	bitlen := bits.Len(uLength)
	k := 1 << uint(bitlen-1)
	if k == n {
		k >>= 1
	}
	return k
}
