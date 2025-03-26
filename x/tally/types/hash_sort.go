package types

import (
	"bytes"
	"crypto/sha256"
	"slices"
)

type HashSortable interface {
	GetSortKey() []byte
}

func HashSort[T HashSortable](originalItems []T, entropy []byte) []T {
	type sortItem struct {
		sortKey  []byte
		original T
	}
	totalItems := len(originalItems)

	sortItems := make([]sortItem, totalItems)

	hasher := sha256.New()
	for i := range totalItems {
		hasher.Reset()

		hasher.Write(originalItems[i].GetSortKey())
		hasher.Write(entropy)

		sortItems[i] = sortItem{
			sortKey:  hasher.Sum(nil),
			original: originalItems[i],
		}
	}

	slices.SortFunc(sortItems, func(a, b sortItem) int {
		return bytes.Compare(a.sortKey, b.sortKey)
	})

	result := make([]T, totalItems)
	for i, item := range sortItems {
		result[i] = item.original
	}

	return result
}
