package utils

import (
	"bytes"
	"fmt"
	"hash"
	"sort"

	"golang.org/x/crypto/sha3"
)

// GetProof returns the merkle proof for the entry at the given index.
func GetProof(entries [][]byte, index int) ([][]byte, error) {
	if index < 0 || index >= len(entries) {
		return nil, fmt.Errorf("index %d is out of bounds", index)
	}
	if len(entries) == 1 {
		return [][]byte{}, nil
	}

	// Hash and sort the leaves.
	sha := sha3.NewLegacyKeccak256()
	leaves := make([][]byte, len(entries))
	for i, entry := range entries {
		leaves[i] = leafHash(sha, entry)
	}
	sort.Slice(leaves, func(i, j int) bool {
		return bytes.Compare(leaves[i], leaves[j]) == -1
	})

	// Build the tree.
	tree := buildTree(sha, leaves)

	// Find the index of the entry in the tree.
	treeIndex := -1
	want := leafHash(sha, entries[index])
	for i := range tree {
		if bytes.Equal(want, tree[i]) {
			treeIndex = i
			break
		}
	}
	if treeIndex == -1 {
		return nil, fmt.Errorf("entry not found in tree")
	}

	var proof [][]byte
	for treeIndex > 0 {
		siblingIndex, err := siblingIndex(treeIndex)
		if err != nil {
			return nil, err
		}
		proof = append(proof, tree[siblingIndex])

		parentIndex, err := parentIndex(treeIndex)
		if err != nil {
			return nil, err
		}
		treeIndex = parentIndex
	}
	return proof, nil
}

func parentIndex(i int) (int, error) {
	if i <= 0 {
		return 0, fmt.Errorf("root has no parent")
	}
	return (i - 1) / 2, nil
}

func siblingIndex(i int) (int, error) {
	if i <= 0 {
		return 0, fmt.Errorf("root has no siblings")
	}
	if i%2 == 0 {
		return i - 1, nil
	}
	return i + 1, nil
}

func VerifyProof(proof [][]byte, root, entry []byte) bool {
	return bytes.Equal(processProof(sha3.NewLegacyKeccak256(), proof, entry), root)
}

func processProof(sha hash.Hash, proof [][]byte, entry []byte) []byte {
	currentHash := leafHash(sha, entry)
	for i := 0; i < len(proof); i++ {
		currentHash = parentHash(sha, currentHash, proof[i])
	}
	return currentHash
}
