package types

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestItem implements HashSortable for testing
type TestItem struct {
	key string
}

func (t TestItem) GetSortKey() []byte {
	return []byte(t.key)
}

type HashSortSuite struct {
	suite.Suite
}

func (s *HashSortSuite) generateItems(numItems int) []TestItem {
	items := make([]TestItem, numItems)
	for i := range numItems {
		items[i] = TestItem{key: string(rune(i))}
	}

	return items
}

func TestHashSortSuite(t *testing.T) {
	suite.Run(t, new(HashSortSuite))
}

func (s *HashSortSuite) TestHashSortEmptyInput() {
	result := HashSort(s.generateItems(0), nil)
	s.Require().Equal(0, len(result), "HashSort() should return an empty slice for empty input")
}

func (s *HashSortSuite) TestHashSortSingleItem() {
	items := s.generateItems(1)
	result := HashSort(items, nil)

	s.Require().Equal(1, len(result), "HashSort() should return slice with length 1 for single input")
	s.Require().Equal(items[0].key, result[0].key, "HashSort() should return same item for single input")
}

func (s *HashSortSuite) TestHashSortSmallInput() {
	items := s.generateItems(10)

	result := HashSort(items, nil)

	s.Require().Equal(10, len(result), "HashSort() should return slice with length 3 for multiple inputs")
}

func (s *HashSortSuite) TestHashSortLargeInput() {
	items := s.generateItems(1000)

	result := HashSort(items, nil)
	s.Require().Equal(1000, len(result), "HashSort() got length %v, want %v", len(result), 1000)

	// Verify all items are present
	seen := make(map[string]bool)
	for _, item := range result {
		seen[item.key] = true
	}
	s.Require().Equal(1000, len(seen), "HashSort() lost some items")
}

func (s *HashSortSuite) TestHashSortDeterminism() {
	// Test that the same items with same keys get sorted consistently
	items := s.generateItems(100)

	result1 := HashSort(items, nil)
	result2 := HashSort(items, nil)

	for i := range len(result1) - 1 {
		s.Require().Equal(result1[i].key, result2[i].key, "HashSort() is not deterministic")
	}

	result3 := HashSort(items, []byte("entropy"))
	result4 := HashSort(items, []byte("entropy"))

	for i := range len(result3) - 1 {
		s.Require().Equal(result3[i].key, result4[i].key, "HashSort() with entropy is not deterministic")
	}
}

func (s *HashSortSuite) TestHashSortEntropy() {
	// With 10 items, probability of all 3 being different at a position is:
	// 1 * (9/10) * (8/10) = 0.72
	items := s.generateItems(10)

	result1 := HashSort(items, nil)
	result2 := HashSort(items, []byte("entropy"))

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, 100)
	result3 := HashSort(items, append([]byte("dr_id"), heightBytes...))

	// Count positions where all 3 results are different
	differentCount := 0
	for i := range len(result1) {
		if result1[i].key != result2[i].key &&
			result2[i].key != result3[i].key &&
			result1[i].key != result3[i].key {
			differentCount++
		}
	}

	// With 10 items, we expect about 7.2 positions to have all different items
	// Setting threshold at 3 to account for random variation while still being statistically significant
	s.Require().GreaterOrEqual(differentCount, 3,
		"Expected at least 3 positions with all different items, got %d", differentCount)
}
