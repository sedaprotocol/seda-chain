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

// ApplyFilter applies the Mode Filter and returns an outlier list.
// (i) If more than 1/3 of reveals are corrupted, a corrupt reveals
// error is returned without an outlier list.
// (ii) Otherwise, a reveal is declared an outlier if it does not
// match the mode value. If less than 2/3 of the reveals are outliers,
// no consensus error is returned along with an outlier list.
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

// ApplyFilter applies the Standard Deviation Filter and returns an
// outlier list.
// (i) If more than 1/3 of reveals are corrupted (i.e. invalid json
// path, invalid bytes, etc.), a corrupt reveals error is returned
// without an outlier list.
// (ii) If the number type is invalid, an error is returned without
// an outlier list.
// (iii) Otherwise, an outlier list is returned. A reveal is declared
// an outlier if it deviates from the median by more than the given
// max sigma. If less than 2/3 of the reveals are outliers, no consensus
// error is returned as well.
func (f FilterStdDev) ApplyFilter(reveals []RevealBody) ([]int, error) {
	dataList, _, err := parseReveals(reveals, f.dataPath)
	if err != nil {
		return nil, err
	}

	outliers, err := f.DetectOutliers(dataList)
	if err != nil {
		return outliers, err
	}
	return outliers, nil
}

func (f FilterStdDev) DetectOutliers(dataList []string) ([]int, error) {
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
		return nil, ErrInvalidNumberType
	}
}

func detectOutliersInteger[T constraints.Integer](dataList []string, maxSigma uint64) ([]int, error) {
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
		return nil, ErrCorruptReveals
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
		return outliers, ErrNoConsensus
	}
	return outliers, nil
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
func (f FilterNone) ApplyFilter(reveals []RevealBody) ([]int, error) {
	return make([]int, len(reveals)), nil
}
