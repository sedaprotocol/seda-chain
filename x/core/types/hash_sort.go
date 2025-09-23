package types

import (
	"bytes"
	"crypto/sha256"
	"slices"
)

type HashSortable interface {
	GetSortKey() []byte
}

// HashSort sorts a given slice of items by the hash of the item's sort key and
// the given entropy. If the entropy is nil, the items are sorted by the item's
// sort key without hashing.
func HashSort[T HashSortable](originalItems []T, entropy []byte) []T {
	type sortItem struct {
		sortKey  []byte
		original T
	}
	totalItems := len(originalItems)

	sortItems := make([]sortItem, totalItems)

	hasher := sha256.New()
	for i := range totalItems {
		if entropy != nil {
			hasher.Reset()
			hasher.Write(originalItems[i].GetSortKey())
			hasher.Write(entropy)

			sortItems[i] = sortItem{
				sortKey:  hasher.Sum(nil),
				original: originalItems[i],
			}
		} else {
			sortItems[i] = sortItem{
				sortKey:  originalItems[i].GetSortKey(),
				original: originalItems[i],
			}
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
