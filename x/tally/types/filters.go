package types

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"slices"
	"strconv"

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
	// an outlier, and a boolean indicating whether consensus in reveal
	// data has been reached.
	ApplyFilter(reveals []RevealBody, errors []bool) ([]bool, bool)
}

type FilterNone struct{}

// NewFilterNone constructs a new FilterNone object.
func NewFilterNone(gasCost uint64, gasMeter *GasMeter) (FilterNone, error) {
	outOfGas := gasMeter.ConsumeTallyGas(gasCost)
	if outOfGas {
		return FilterNone{}, ErrOutofTallyGas
	}
	return FilterNone{}, nil
}

// FilterNone declares all reveals as non-outliers.
func (f FilterNone) ApplyFilter(reveals []RevealBody, _ []bool) ([]bool, bool) {
	return make([]bool, len(reveals)), true
}

type FilterMode struct {
	dataPath          string // JSON path to reveal data
	replicationFactor uint16
}

// NewFilterMode constructs a new FilerMode object given a filter
// input.
// Mode filter input looks as follows:
// 0             1                  9       9+data_path_length
// | filter_type | data_path_length |   data_path   |
func NewFilterMode(input []byte, gasCostMultiplier uint64, replicationFactor uint16, gasMeter *GasMeter) (FilterMode, error) {
	outOfGas := gasMeter.ConsumeTallyGas(gasCostMultiplier * uint64(replicationFactor))
	if outOfGas {
		return FilterMode{}, ErrOutofTallyGas
	}

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
	filter.replicationFactor = replicationFactor
	return filter, nil
}

// ApplyFilter applies the Mode Filter and returns an outlier list.
// A reveal is declared an outlier if it does not match the mode value.
// If less than 2/3 of the reveals are non-outliers, "no consensus"
// error is returned along with an outlier list.
func (f FilterMode) ApplyFilter(reveals []RevealBody, errors []bool) ([]bool, bool) {
	dataList, dataAttrs := parseReveals(reveals, f.dataPath, errors)

	outliers := make([]bool, len(reveals))
	for i, r := range dataList {
		if dataAttrs.freqMap[r] != dataAttrs.maxFreq {
			outliers[i] = true
		}
	}
	if dataAttrs.maxFreq*3 < int(f.replicationFactor)*2 {
		return outliers, false
	}
	return outliers, true
}

type FilterStdDev struct {
	maxSigma          Sigma
	dataPath          string // JSON path to reveal data
	filterFunc        func(dataList []any, maxSigma Sigma, errors []bool, replicationFactor uint16) ([]bool, bool)
	replicationFactor uint16
}

// NewFilterStdDev constructs a new FilterStdDev object given a
// filter input.
// Standard deviation filter input looks as follows:
// 0             1           9             10                 18 18+json_path_length
// | filter_type | max_sigma | number_type | json_path_length | json_path |
func NewFilterStdDev(input []byte, gasCostMultiplier uint64, replicationFactor uint16, gasMeter *GasMeter) (FilterStdDev, error) {
	outOfGas := gasMeter.ConsumeTallyGas(gasCostMultiplier * uint64(replicationFactor))
	if outOfGas {
		return FilterStdDev{}, ErrOutofTallyGas
	}

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
		filter.filterFunc = detectOutliersSignedInteger[int32]
	case 0x01: // Int64
		filter.filterFunc = detectOutliersSignedInteger[int64]
	case 0x02: // Uint32
		filter.filterFunc = detectOutliersUnsignedInt[uint32]
	case 0x03: // Uint64
		filter.filterFunc = detectOutliersUnsignedInt[uint64]
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
	filter.replicationFactor = replicationFactor
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
func (f FilterStdDev) ApplyFilter(reveals []RevealBody, errors []bool) ([]bool, bool) {
	dataList, _ := parseReveals(reveals, f.dataPath, errors)
	return f.filterFunc(dataList, f.maxSigma, errors, f.replicationFactor)
}

func detectOutliersSignedInteger[T constraints.Integer](dataList []any, maxSigma Sigma, errors []bool, replicationFactor uint16) ([]bool, bool) {
	nums := make([]T, 0, len(dataList))
	corruptQueue := make([]int, 0, len(dataList)) // queue of corrupt indices in dataList
	for i, data := range dataList {
		if data == nil {
			errors[i] = true
			corruptQueue = append(corruptQueue, i)
			continue
		}
		num, ok := data.(int64)
		if !ok {
			errors[i] = true
			corruptQueue = append(corruptQueue, i)
			continue
		}
		nums = append(nums, T(num))
	}

	// Construct outliers list.
	outliers := make([]bool, len(dataList))
	if len(nums) == 0 {
		return outliers, false
	}
	median := findMedian(nums)
	var numsInd, nonOutlierCount int
	for i := range outliers {
		if len(corruptQueue) > 0 && i == corruptQueue[0] {
			outliers[i] = true
			corruptQueue = corruptQueue[1:]
		} else {
			if median.IsWithinSigma(nums[numsInd], maxSigma) {
				nonOutlierCount++
			} else {
				outliers[i] = true
			}
			numsInd++
		}
	}

	// If less than 2/3 of the numbers fall within max sigma range
	// from the median, there is no consensus in reveal data.
	if nonOutlierCount*3 < int(replicationFactor)*2 {
		return outliers, false
	}
	return outliers, true
}

func detectOutliersUnsignedInt[T constraints.Unsigned](dataList []any, maxSigma Sigma, errors []bool, replicationFactor uint16) ([]bool, bool) {
	nums := make([]T, 0, len(dataList))
	corruptQueue := make([]int, 0, len(dataList)) // queue of corrupt indices in dataList
	for i, data := range dataList {
		if data == nil {
			errors[i] = true
			corruptQueue = append(corruptQueue, i)
			continue
		}
		var num T
		num1, ok := data.(int64)
		if !ok {
			jsonNum, ok := data.(json.Number)
			if !ok {
				errors[i] = true
				corruptQueue = append(corruptQueue, i)
				continue
			}

			num2, err := strconv.ParseUint(jsonNum.String(), 10, 64)
			if err != nil {
				errors[i] = true
				corruptQueue = append(corruptQueue, i)
				continue
			}
			num = T(num2)
		} else {
			num = T(num1)
		}
		nums = append(nums, num)
	}

	// Construct outliers list.
	outliers := make([]bool, len(dataList))
	if len(nums) == 0 {
		return outliers, false
	}
	median := findMedian(nums)
	var numsInd, nonOutlierCount int
	for i := range outliers {
		if len(corruptQueue) > 0 && i == corruptQueue[0] {
			outliers[i] = true
			corruptQueue = corruptQueue[1:]
		} else {
			if median.IsWithinSigma(nums[numsInd], maxSigma) {
				nonOutlierCount++
			} else {
				outliers[i] = true
			}
			numsInd++
		}
	}

	// If less than 2/3 of the numbers fall within max sigma range
	// from the median, there is no consensus in reveal data.
	if nonOutlierCount*3 < int(replicationFactor)*2 {
		return outliers, false
	}
	return outliers, true
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
