package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"

	"golang.org/x/exp/constraints"
)

type Filter interface {
	// ApplyFilter applies filter and returns an outlier list,
	// a consensus boolean, and an error if applicable.
	ApplyFilter(reveals []RevealBody) ([]int, bool, error)
}

type FilterMode struct {
	dataPath string // JSON path to reveal data
}

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

// ApplyFilter takes in a list of reveals. It returns an outlier list,
// a consensus boolean, and an error if applicable.
// Value of 1 at index i in the outlier list indicates that the i-th
// reveal is an outlier. Value of 0 indicates a non-outlier reveal.
// A reveal is an outlier if it has less than or equal to 2/3 frequency.
// The consensus boolean is true if one of the following criteria is met:
// 1. More than 1/3 of the reveals are corrupted (non-zero exit code,
// invalid data path, etc.)
// 2. More than 2/3 of the reveals are identical.
func (f FilterMode) ApplyFilter(reveals []RevealBody) ([]int, bool, error) {
	dataList, outliers, consensus, freq, maxFreq := ParseReveals(reveals, f.dataPath)
	if dataList == nil {
		return outliers, consensus, nil
	}

	outliers = allOutlierList(len(dataList))

	// If less than 2/3 of the reveals match the max frequency,
	// there is no consensus.
	if maxFreq*3 < len(reveals)*2 {
		return outliers, false, nil
	}

	for i, r := range dataList {
		if freq[r] == maxFreq {
			outliers[i] = 0
		}
	}
	return outliers, true, nil
}

type FilterStdDev struct {
	maxSigma   uint64 // 10^6 precision
	numberType byte
	dataPath   string // JSON path to reveal data
}

// Standard deviation filter input looks as follows:
// 0             1           9             10                 18 18+json_path_length
// | filter_type | max_sigma | number_type | json_path_length | json_path |
func NewFilterStdDev(input []byte) (FilterStdDev, error) {
	var filter FilterStdDev
	if len(input) < 18 {
		return filter, ErrInvalidFilterInputLen.Wrapf("%d < %d", len(input), 18)
	}

	var maxSigma uint64
	err := binary.Read(bytes.NewReader(input[1:9]), binary.BigEndian, &maxSigma)
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

// TODO Add comments
func (f FilterStdDev) ApplyFilter(reveals []RevealBody) ([]int, bool, error) {
	dataList, outliers, consensus, _, _ := ParseReveals(reveals, f.dataPath)
	if dataList == nil {
		return outliers, consensus, nil
	}

	outliers, consensus, err := f.DetectOutliers(dataList)
	if err != nil {
		return outliers, consensus, err
	}
	return outliers, true, nil
}

func (f FilterStdDev) DetectOutliers(dataList []string) ([]int, bool, error) {
	switch f.numberType {
	case 0x00: // Int32
		return detectOutliersInteger[int32](dataList, f.maxSigma)
	case 0x01: // Int64
		return detectOutliersInteger[int64](dataList, f.maxSigma)
	case 0x02: // Uint32
		return detectOutliersInteger[uint32](dataList, f.maxSigma)
	case 0x03: // Uint64
		return detectOutliersInteger[uint64](dataList, f.maxSigma)
	// TODO: Support other types
	default:
		return nil, false, fmt.Errorf("invalid number type")
	}
}

// detectOutliersInteger converts a list of data in string to a list of
// numbers.
// TODO elaborate comment
func detectOutliersInteger[T constraints.Integer](dataList []string, maxSigma uint64) ([]int, bool, error) {
	length := len(dataList)
	if length == 0 {
		panic("zero data list length") // TODO should never end up here?
	}

	var corruptCount int
	var z T
	rt := reflect.TypeOf(z)
	numbers := make([]T, length)
	for i, data := range dataList {
		dataBytes, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			corruptCount++
			continue
		}

		switch rt.Kind() {
		case reflect.Uint64:
			// TODO length check?
			data := binary.BigEndian.Uint64(dataBytes)
			numbers[i] = T(data)

		// TODO: Support other types
		default:
			panic("invalid number type") // TODO should never end up here
		}
	}

	// If more than 1/3 of the reveals are corrupted,
	// we reach consensus that the reveals are unusable.
	if corruptCount*3 > length {
		return allOutlierList(length), true, nil
	}

	// Sort and find median.
	slices.Sort(numbers)

	var median T
	var medianHalf bool // if true, median must be corrected by adding 0.5
	if length%2 == 1 {
		median = numbers[length/2]
	} else {
		mid := length / 2
		median = (numbers[mid-1] + numbers[mid]) / 2
		if (numbers[mid-1]%2 == 0 && numbers[mid]%2 == 1) ||
			(numbers[mid-1]%2 == 1 && numbers[mid]%2 == 0) {
			medianHalf = true
		}
	}

	// Identify outliers and keep their count.
	outliers := make([]int, len(numbers))
	var nonOutlierCount int
	for i, num := range numbers {
		if isOutlier(maxSigma, num, median, medianHalf) {
			outliers[i] = 1
		}
		nonOutlierCount++
	}

	// If less than 2/3 of the numbers fall within max sigma range
	// from the median, there is no consensus.
	if nonOutlierCount*3 < len(numbers)*2 {
		return nil, false, nil
	}
	return outliers, true, nil
}

func isOutlier[T constraints.Integer](maxSigma uint64, num, median T, medianHalf bool) bool {
	var diff uint64
	if median > num {
		diff = uint64(median - num) // + 0.5 if medianHalf=true
	} else if median < num {
		diff = uint64(num - median) // - 0.5 if medianHalf=true
		if medianHalf {
			diff--
		}
	} else {
		return false
	}

	if diff > (maxSigma / 1e6) {
		return true
	} else if medianHalf && diff == (maxSigma/1e6) {
		// Means that diff = int(maxSigma) + 0.5
		// so now check that maxSigma's decimal part is > 0.5
		// by checking that last 6 digits of maxSigma
		return maxSigma%1e6 > 5e6
	}
	return false
}

type FilterNone struct{}

func NewFilterNone(_ []byte) (FilterNone, error) {
	return FilterNone{}, nil
}

// FilterNone declares all reveals as non-outliers with consensus.
func (f FilterNone) ApplyFilter(reveals []RevealBody) ([]int, bool, error) {
	return make([]int, len(reveals)), true, nil
}
