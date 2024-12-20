package types

import (
	"bytes"
	"encoding/binary"
	"slices"

	"golang.org/x/exp/constraints"
)

var (
	_ Filter = &FilterNone{}
	_ Filter = &FilterMode{}
	_ Filter = &FilterStdDev{}
)

type Filter interface {
	// ApplyFilter takes in a list of reveals and returns an outlier
	// list, whose value at index i indicates whether i-th reveal is
	// an outlier. Value of 1 indicates an outlier, and value of 0
	// indicates a non-outlier reveal.
	ApplyFilter(reveals []RevealBody) ([]int, error)
	// GasCost returns the cost of the filter in terms of gas amount.
	GasCost() uint64
}

type FilterNone struct {
	gasCost uint64
}

// NewFilterNone constructs a new FilterNone object.
func NewFilterNone(gasCost uint64) FilterNone {
	return FilterNone{gasCost: gasCost}
}

// FilterNone declares all reveals as non-outliers, unless reveals are
// empty or corrupt.
func (f FilterNone) ApplyFilter(reveals []RevealBody) ([]int, error) {
	if len(reveals) == 0 {
		return nil, ErrEmptyReveals // TODO remove since unreachable?
	}

	var corruptCount int
	for _, r := range reveals {
		if r.ExitCode != 0 {
			corruptCount++
			continue
		}
	}
	if corruptCount*3 > len(reveals) {
		return nil, ErrCorruptReveals
	}

	return make([]int, len(reveals)), nil
}

func (f FilterNone) GasCost() uint64 {
	return f.gasCost
}

type FilterMode struct {
	dataPath string // JSON path to reveal data
	gasCost  uint64
}

// NewFilterMode constructs a new FilerMode object given a filter
// input.
// Mode filter input looks as follows:
// 0             1                  9       9+data_path_length
// | filter_type | data_path_length |   data_path   |
func NewFilterMode(input []byte, gasCostMultiplier uint64, replicationFactor uint16) (FilterMode, error) {
	var filter FilterMode
	if len(input) < 9 {
		return filter, ErrFilterInputTooShort.Wrapf("%d < %d", len(input), 9)
	}

	var pathLen uint64
	err := binary.Read(bytes.NewReader(input[1:9]), binary.BigEndian, &pathLen)
	if err != nil {
		return filter, err
	}

	path := input[9:]
	if len(path) != int(pathLen) /* #nosec G115 */ {
		return filter, ErrInvalidPathLen.Wrapf("expected: %d got: %d", int(pathLen), len(path)) // #nosec G115
	}
	filter.dataPath = string(path)
	filter.gasCost = gasCostMultiplier * uint64(replicationFactor)
	return filter, nil
}

func (f FilterMode) GasCost() uint64 {
	return f.gasCost
}

// ApplyFilter applies the Mode Filter and returns an outlier list.
// (i) If more than 1/3 of reveals are corrupted, a corrupt reveals
// error is returned without an outlier list.
// (ii) Otherwise, a reveal is declared an outlier if it does not
// match the mode value. If less than 2/3 of the reveals are non-outliers,
// "no consensus" error is returned along with an outlier list.
func (f FilterMode) ApplyFilter(reveals []RevealBody) ([]int, error) {
	dataList, dataAttrs, err := parseReveals(reveals, f.dataPath)
	if err != nil {
		return nil, err
	}

	outliers := make([]int, len(reveals))
	for i, r := range dataList {
		if dataAttrs.freqMap[r] != dataAttrs.maxFreq {
			outliers[i] = 1
		}
	}
	if dataAttrs.maxFreq*3 < len(reveals)*2 {
		return outliers, ErrNoConsensus
	}
	return outliers, nil
}

type FilterStdDev struct {
	maxSigma   Sigma
	dataPath   string // JSON path to reveal data
	filterFunc func(dataList []any, maxSigma Sigma) ([]int, error)
	gasCost    uint64
}

// NewFilterStdDev constructs a new FilterStdDev object given a
// filter input.
// Standard deviation filter input looks as follows:
// 0             1           9             10                 18 18+json_path_length
// | filter_type | max_sigma | number_type | json_path_length | json_path |
func NewFilterStdDev(input []byte, gasCostMultiplier uint64, replicationFactor uint16) (FilterStdDev, error) {
	var filter FilterStdDev
	if len(input) < 18 {
		return filter, ErrFilterInputTooShort.Wrapf("%d < %d", len(input), 18)
	}

	maxSigma, err := NewSigma(input[1:9])
	if err != nil {
		return filter, err
	}
	filter.maxSigma = maxSigma

	switch input[9] {
	case 0x00: // Int32
		filter.filterFunc = detectOutliersInteger[int32]
	case 0x01: // Int64
		filter.filterFunc = detectOutliersInteger[int64]
	case 0x02: // Uint32
		filter.filterFunc = detectOutliersInteger[uint32]
	case 0x03: // Uint64
		filter.filterFunc = detectOutliersInteger[uint64]
	default:
		return filter, ErrInvalidNumberType
	}

	var pathLen uint64
	err = binary.Read(bytes.NewReader(input[10:18]), binary.BigEndian, &pathLen)
	if err != nil {
		return filter, err
	}

	path := input[18:]
	if len(path) != int(pathLen) /* #nosec G115 */ {
		return filter, ErrInvalidPathLen.Wrapf("expected: %d got: %d", int(pathLen), len(path)) // #nosec G115
	}
	filter.dataPath = string(path)
	filter.gasCost = gasCostMultiplier * uint64(replicationFactor)
	return filter, nil
}

// ApplyFilter applies the Standard Deviation Filter and returns an
// outlier list.
// (i) If more than 1/3 of reveals are corrupted (i.e. invalid json
// path, invalid bytes, etc.), a corrupt reveals error is returned
// without an outlier list.
// (ii) If the number type is invalid, an error is returned without
// an outlier list.
// (iii) Otherwise, an outlier list is returned. A reveal is declared
// an outlier if it deviates from the median by more than the given
// max sigma. If less than 2/3 of the reveals are non-outliers, "no
// consensus" error is returned as well.
func (f FilterStdDev) ApplyFilter(reveals []RevealBody) ([]int, error) {
	dataList, _, err := parseReveals(reveals, f.dataPath)
	if err != nil {
		return nil, err
	}
	return f.filterFunc(dataList, f.maxSigma)
}

func (f FilterStdDev) GasCost() uint64 {
	return f.gasCost
}

func detectOutliersInteger[T constraints.Integer](dataList []any, maxSigma Sigma) ([]int, error) {
	nums := make([]T, 0, len(dataList))
	corruptQueue := make([]int, 0, len(dataList)) // queue of corrupt indices in dataList
	for i, data := range dataList {
		if data == nil {
			corruptQueue = append(corruptQueue, i)
			continue
		}
		num, ok := data.(int64)
		if !ok {
			corruptQueue = append(corruptQueue, i)
			continue
		}
		nums = append(nums, T(num))
	}

	// If more than 1/3 of the reveals are corrupted,
	// return corrupt reveals error.
	if len(corruptQueue)*3 > len(dataList) {
		return nil, ErrCorruptReveals
	}

	// Construct outliers list.
	median := findMedian(nums)
	outliers := make([]int, len(dataList))
	var numsInd, nonOutlierCount int
	for i := range outliers {
		if len(corruptQueue) > 0 && i == corruptQueue[0] {
			outliers[i] = 1
			corruptQueue = corruptQueue[1:]
		} else {
			if median.IsWithinSigma(nums[numsInd], maxSigma) {
				nonOutlierCount++
			} else {
				outliers[i] = 1
			}
			numsInd++
		}
	}

	// If less than 2/3 of the numbers fall within max sigma range
	// from the median, there is no consensus.
	if nonOutlierCount*3 < len(nums)*2 {
		return outliers, ErrNoConsensus
	}
	return outliers, nil
}

// findMedian returns the median of a given slice of integers.
// It makes a copy of the slice to leave the given slice intact.
func findMedian[T constraints.Integer](nums []T) *HalfStepInt[T] {
	length := len(nums)
	numsSorted := make([]T, length)
	copy(numsSorted, nums)
	slices.Sort(numsSorted)

	median := new(HalfStepInt[T])
	mid := length / 2
	if length%2 == 1 {
		median.Mid(numsSorted[mid], numsSorted[mid])
	} else {
		median.Mid(numsSorted[mid-1], numsSorted[mid])
	}
	return median
}
