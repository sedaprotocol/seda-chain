package types

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/big"
	"slices"
)

var (
	_ Filter = &FilterNone{}
	_ Filter = &FilterMode{}
	_ Filter = &FilterMAD{}
)

type Filter interface {
	// ApplyFilter takes in a list of reveals and returns an outlier
	// list, whose value at index i indicates whether i-th reveal is
	// an outlier, and a boolean indicating whether consensus in reveal
	// data has been reached.
	ApplyFilter(reveals []Reveal, errors []bool) ([]bool, bool)
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
func (f FilterNone) ApplyFilter(reveals []Reveal, _ []bool) ([]bool, bool) {
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
func (f FilterMode) ApplyFilter(reveals []Reveal, errors []bool) ([]bool, bool) {
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

// FilterMAD implements a Median Absolute Deviation filter.
type FilterMAD struct {
	sigmaMultiplier   SigmaMultiplier
	dataPath          string // JSON path to reveal data
	replicationFactor uint16
	// The maximum and minimum values that can be represented by the number type as specified by the requestor
	maxNumber *big.Int
	minNumber *big.Int
}

var (
	minUint       = big.NewInt(0)
	minInt32      = big.NewInt(0).SetInt64(math.MinInt32)
	maxInt32      = big.NewInt(0).SetInt64(math.MaxInt32)
	maxUint32     = big.NewInt(0).SetUint64(math.MaxUint32)
	minInt64      = big.NewInt(0).SetInt64(math.MinInt64)
	maxInt64      = big.NewInt(0).SetInt64(math.MaxInt64)
	maxUint64     = big.NewInt(0).SetUint64(math.MaxUint64)
	minInt128, _  = big.NewInt(0).SetString("-170141183460469231731687303715884105728", 10)
	maxInt128, _  = big.NewInt(0).SetString("170141183460469231731687303715884105727", 10)
	maxUint128, _ = big.NewInt(0).SetString("340282366920938463463374607431768211455", 10)
	minInt256, _  = big.NewInt(0).SetString("-57896044618658097711785492504343953926634992332820282019728792003956564819968", 10)
	maxInt256, _  = big.NewInt(0).SetString("57896044618658097711785492504343953926634992332820282019728792003956564819967", 10)
	maxUint256, _ = big.NewInt(0).SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
)

// NewFilterMAD constructs a Median Absolute Deviation filter given a filter
// input in following format:
// 0             1           9             10                 18 18+json_path_length
// | filter_type | max_sigma | number_type | json_path_length | json_path |
func NewFilterMAD(input []byte, gasCostMultiplier uint64, replicationFactor uint16, gasMeter *GasMeter) (FilterMAD, error) {
	outOfGas := gasMeter.ConsumeTallyGas(gasCostMultiplier * uint64(replicationFactor))
	if outOfGas {
		return FilterMAD{}, ErrOutofTallyGas
	}

	var filter FilterMAD
	if len(input) < 18 {
		return filter, ErrFilterInputTooShort.Wrapf("%d < %d", len(input), 18)
	}

	sigmaMultiplier, err := NewSigmaMultiplier(input[1:9])
	if err != nil {
		return filter, err
	}
	filter.sigmaMultiplier = sigmaMultiplier

	switch input[9] {
	case 0x00: // 32-bit signed integer
		filter.maxNumber = maxInt32
		filter.minNumber = minInt32
	case 0x01: // 32-bit unsigned integer
		filter.maxNumber = maxUint32
		filter.minNumber = minUint
	case 0x02: // 64-bit signed integer
		filter.maxNumber = maxInt64
		filter.minNumber = minInt64
	case 0x03: // 64-bit unsigned integer
		filter.maxNumber = maxUint64
		filter.minNumber = minUint
	case 0x04: // 128-bit signed integer
		filter.maxNumber = maxInt128
		filter.minNumber = minInt128
	case 0x05: // 128-bit unsigned integer
		filter.maxNumber = maxUint128
		filter.minNumber = minUint
	case 0x06: // 256-bit signed integer
		filter.maxNumber = maxInt256
		filter.minNumber = minInt256
	case 0x07: // 256-bit unsigned integer
		filter.maxNumber = maxUint256
		filter.minNumber = minUint
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

// ApplyFilter applies the Median Absolute Deviation Filter and returns an
// outlier list. A reveal is declared an outlier if it deviates from the median
// by more than the median absolute deviation multiplied by the given sigma
// multiplier value.
func (f FilterMAD) ApplyFilter(reveals []Reveal, errors []bool) ([]bool, bool) {
	dataList, _ := parseReveals(reveals, f.dataPath, errors)
	return detectOutliersBigInt(dataList, f.sigmaMultiplier, errors, f.replicationFactor, f.minNumber, f.maxNumber)
}

func detectOutliersBigInt(dataList []string, sigmaMultiplier SigmaMultiplier, errors []bool, replicationFactor uint16, minNumber *big.Int, maxNumber *big.Int) ([]bool, bool) {
	// Parse data list in strings into big Ints.
	sum := new(big.Int)
	nums := make([]*big.Int, 0, len(dataList))
	corruptQueue := make([]int, 0, len(dataList)) // queue of corrupt indices in dataList
	for i, data := range dataList {
		if data == "" {
			errors[i] = true
			corruptQueue = append(corruptQueue, i)
			continue
		}

		num := new(big.Int)
		_, ok := num.SetString(data, 10)
		if !ok || num.Cmp(minNumber) < 0 || num.Cmp(maxNumber) > 0 {
			errors[i] = true
			corruptQueue = append(corruptQueue, i)
			continue
		}
		nums = append(nums, num)
		sum.Add(sum, num)
	}

	// Construct outliers list.
	outliers := make([]bool, len(dataList))
	if len(nums) == 0 {
		return outliers, false
	}

	median := getMedian(nums)
	absDevs := make([]*big.Rat, len(nums))
	for i, num := range nums {
		absDevs[i] = new(big.Rat).Abs(new(big.Rat).Sub(new(big.Rat).SetInt(num), median))
	}
	mad := getMedianRat(absDevs)
	maxDev := new(big.Rat).Mul(sigmaMultiplier.BigRat(), mad)

	// Fill out the outliers list.
	var numsInd, nonOutlierCount int
	for i := range outliers {
		if len(corruptQueue) > 0 && i == corruptQueue[0] {
			outliers[i] = true
			corruptQueue = corruptQueue[1:]
		} else {
			if isWithinMaxDev(nums[numsInd], median, maxDev) {
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

// isWithinMaxDev returns true if the given number is within the given
// deviation amount from the mean.
func isWithinMaxDev(num *big.Int, median, maxDev *big.Rat) bool {
	diff := new(big.Rat).Sub(new(big.Rat).SetInt(num), median)
	return maxDev.Cmp(new(big.Rat).Abs(diff)) >= 0
}

// getMedian returns the median of the given list of big.Ints.
func getMedian(nums []*big.Int) *big.Rat {
	sortNums := make([]*big.Int, len(nums))
	copy(sortNums, nums)
	slices.SortFunc(sortNums, func(a, b *big.Int) int {
		return a.Cmp(b)
	})

	l := len(sortNums)
	median := new(big.Rat)
	if l%2 == 0 {
		sum := new(big.Int).Add(sortNums[l/2-1], sortNums[l/2])
		median.SetFrac(sum, big.NewInt(2))
	} else {
		median.SetInt(sortNums[l/2])
	}
	return median
}

func getMedianRat(nums []*big.Rat) *big.Rat {
	sortNums := make([]*big.Rat, len(nums))
	copy(sortNums, nums)
	slices.SortFunc(sortNums, func(a, b *big.Rat) int {
		return a.Cmp(b)
	})

	l := len(sortNums)
	median := new(big.Rat)
	if l%2 == 0 {
		median.Add(sortNums[l/2-1], sortNums[l/2])
		median.Quo(median, big.NewRat(2, 1))
	} else {
		median.Set(sortNums[l/2])
	}
	return median
}
