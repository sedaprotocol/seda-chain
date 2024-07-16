package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"reflect"
	"slices"

	"golang.org/x/exp/constraints"
)

type Filter interface {
	// ApplyFilter takes in a list of reveals and returns an outlier
	// list, whose value at index i indicates whether i-th reveal is
	// an outlier. Value of 1 indicates an outlier, and value of 0
	// indicates a non-outlier reveal.
	ApplyFilter(reveals []RevealBody) ([]int, error)
}

type FilterNone struct{}

// NewFilterNone constructs a new FilterNone object.
func NewFilterNone(_ []byte) (FilterNone, error) {
	return FilterNone{}, nil
}

// FilterNone declares all reveals as non-outliers.
func (f FilterNone) ApplyFilter(reveals []RevealBody) ([]int, error) {
	return make([]int, len(reveals)), nil
}

type FilterMode struct {
	dataPath string // JSON path to reveal data
}

// NewFilterMode constructs a new FilerMode object given a filter
// input.
// Mode filter input looks as follows:
// 0             1                  9       9+data_path_length
// | filter_type | data_path_length |   data_path   |
func NewFilterMode(input []byte) (FilterMode, error) {
	var filter FilterMode
	if len(input) < 9 {
		return filter, ErrInvalidFilterInputLen.Wrapf("%d < %d", len(input), 9)
	}

	var pathLen uint64
	err := binary.Read(bytes.NewReader(input[1:9]), binary.BigEndian, &pathLen)
	if err != nil {
		return filter, err
	}

	path := input[9:]
	if len(path) != int(pathLen) {
		return filter, ErrInvalidPathLen.Wrapf("expected: %d got: %d", int(pathLen), len(path))
	}
	filter.dataPath = string(path)
	return filter, nil
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
	numberType byte
	dataPath   string // JSON path to reveal data
}

// NewFilterStdDev constructs a new FilterStdDev object given a
// filter input.
// Standard deviation filter input looks as follows:
// 0             1           9             10                 18 18+json_path_length
// | filter_type | max_sigma | number_type | json_path_length | json_path |
func NewFilterStdDev(input []byte) (FilterStdDev, error) {
	var filter FilterStdDev
	if len(input) < 18 {
		return filter, ErrInvalidFilterInputLen.Wrapf("%d < %d", len(input), 18)
	}

	maxSigma, err := NewSigma(input[1:9])
	if err != nil {
		return filter, err
	}
	filter.maxSigma = maxSigma

	filter.numberType = input[9]

	var pathLen uint64
	err = binary.Read(bytes.NewReader(input[10:18]), binary.BigEndian, &pathLen)
	if err != nil {
		return filter, err
	}

	path := input[18:]
	if len(path) != int(pathLen) {
		return filter, ErrInvalidPathLen.Wrapf("expected: %d got: %d", int(pathLen), len(path))
	}
	filter.dataPath = string(path)
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

	outliers, err := f.detectOutliers(dataList)
	if err != nil {
		return outliers, err
	}
	return outliers, nil
}

func (f FilterStdDev) detectOutliers(dataList []string) ([]int, error) {
	switch f.numberType {
	case 0x00: // Int32
		return detectOutliersInteger[int32](dataList, f.maxSigma)
	case 0x01: // Int64
		return detectOutliersInteger[int64](dataList, f.maxSigma)
	case 0x02: // Uint32
		return detectOutliersInteger[uint32](dataList, f.maxSigma)
	case 0x03: // Uint64
		return detectOutliersInteger[uint64](dataList, f.maxSigma)
	default:
		return nil, ErrInvalidNumberType
	}
}

func detectOutliersInteger[T constraints.Integer](dataList []string, maxSigma Sigma) ([]int, error) {
	nums := make([]T, 0, len(dataList))
	corruptQueue := make([]int, 0, len(dataList)) // queue of corrupt indices in dataList
	typeSize := int(reflect.TypeOf(new(T)).Size())
	for i, data := range dataList {
		if data == "" {
			corruptQueue = append(corruptQueue, i)
			continue
		}
		bz, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			corruptQueue = append(corruptQueue, i)
			continue
		}
		if len(bz) != typeSize {
			corruptQueue = append(corruptQueue, i)
			continue
		}

		var num T
		err = binary.Read(bytes.NewBuffer(bz), binary.BigEndian, &num)
		if err != nil {
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
	median, medianHalf := findMedian(nums)
	outliers := make([]int, len(dataList))
	var numsInd, nonOutlierCount int
	for i := range outliers {
		if len(corruptQueue) > 0 && i == corruptQueue[0] {
			outliers[i] = 1
			corruptQueue = corruptQueue[1:]
		} else {
			if isOutlier(maxSigma, nums[numsInd], median, medianHalf) {
				outliers[i] = 1
			} else {
				nonOutlierCount++
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

// findMedian sorts a given list of integers to find the median.
// The median is represented by two variables of types T and bool,
// respectively. T represents the integer part of the median and
// the bool, if set to true, represents 0.5 step addition to T.
func findMedian[T constraints.Integer](nums []T) (T, bool) {
	// Sort and find median.
	length := len(nums)
	numsSorted := make([]T, length)
	copy(numsSorted, nums)
	slices.Sort(numsSorted)

	var median T
	var medianHalf bool // if true, median must be corrected by adding 0.5
	if length%2 == 1 {
		median = numsSorted[length/2]
	} else {
		mid := length / 2
		median = (numsSorted[mid-1] + numsSorted[mid]) / 2
		// Consider the case when the median contains fractional part
		// that has been truncated by the integer division.
		if (numsSorted[mid-1]%2 == 0 && numsSorted[mid]%2 == 1) ||
			(numsSorted[mid-1]%2 == 1 && numsSorted[mid]%2 == 0) {
			medianHalf = true
		}
	}
	return median, medianHalf
}

func isOutlier[T constraints.Integer](maxSigma Sigma, num, median T, medianHalf bool) bool {
	var diff uint64
	switch {
	case median > num:
		diff = uint64(median - num)
	case median < num:
		diff = uint64(num - median)
		if medianHalf {
			diff--
		}
	case median == num:
		return false
	}

	if diff > maxSigma.WholeNumber() {
		return true
	} else if medianHalf && diff == maxSigma.WholeNumber() {
		// If we reach here, it means that diff = int(maxSigma) + 0.5.
		// Therefore, we check that maxSigma's fractional part is
		// smaller than 0.5.
		return maxSigma.FractionalPart() < 5e5
	}
	return false
}
