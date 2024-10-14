package utils

import (
	"bytes"
	"hash"
	"sort"

	"golang.org/x/crypto/sha3"
)

// RootFromEntries computes the root of a merkle tree whose leaves are
// the hashes of the given entries.
//   - It uses Keccak-256 hashing algorithm.
//   - The leaves are sorted.
//   - Each hash pair is sorted to make proofs more succinct.
func RootFromEntries(entries [][]byte) []byte {
	// Hash the entries and sort the leaves.
	sha := sha3.NewLegacyKeccak256()
	leaves := make([][]byte, len(entries))
	for i, entry := range entries {
		leaves[i] = leafHash(sha, entry)
	}
	sort.Slice(leaves, func(i, j int) bool {
		return bytes.Compare(leaves[i], leaves[j]) == -1
	})

	return buildTree(sha, leaves)[0]
}

// RootFromLeaves computes the root of a merkle tree with the given
// leaves.
func RootFromLeaves(leaves [][]byte) []byte {
	// Use an empty hash if a provided leaf is empty.
	sha := sha3.NewLegacyKeccak256()
	for i := range leaves {
		if len(leaves[i]) == 0 {
			leaves[i] = emptyHash(sha)
		}
	}
	// Sort the leaves.
	sort.Slice(leaves, func(i, j int) bool {
		return bytes.Compare(leaves[i], leaves[j]) == -1
	})

	return buildTree(sha, leaves)[0]
}

// buildTree builds and returns a merkle tree from the given leaves.
// The tree root is placed at the first index. If no leaves are
// provided, the returned tree is a single node with an empty hash.
func buildTree(sha hash.Hash, leaves [][]byte) [][]byte {
	if len(leaves) == 0 {
		return [][]byte{emptyHash(sha)}
	}

	tree := make([][]byte, 2*len(leaves)-1)
	for i, leaf := range leaves {
		tree[len(tree)-1-i] = leaf
	}
	for i := len(tree) - 1 - len(leaves); i >= 0; i-- {
		tree[i] = parentHash(sha, tree[2*i+1], tree[2*i+2])
	}
	return tree
}

// parentHash computes the parent's hash given its two children a and b
// that have already been hashed.
func parentHash(sha hash.Hash, a, b []byte) []byte {
	if bytes.Compare(a, b) == -1 {
		return nodeHash(sha, a, b)
	}
	return nodeHash(sha, b, a)
}

// emptyHash returns hash(<empty>)
func emptyHash(s hash.Hash) []byte {
	s.Reset()
	s.Write([]byte{})
	return s.Sum(nil)
}

// leafHash returns hash(leaf)
func leafHash(s hash.Hash, leaf []byte) []byte {
	s.Reset()
	s.Write(leaf)
	return s.Sum(nil)
}

// nodeHash returns hash(left || right)
func nodeHash(s hash.Hash, left []byte, right []byte) []byte {
	s.Reset()
	s.Write(left)
	s.Write(right)
	return s.Sum(nil)
}
