package keeper

import (
	"errors"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
	"github.com/tidwall/gjson"
)

const (
	filterNone   byte = 0x00
	filterMode   byte = 0x01
	filterStdDev byte = 0x02
)

// ApplyFilter processes filter of the type specified in the first byte of
// consensus filter. It returns an outlier list, which is a boolean list where
// true at index i means that the reveal at index i is an outlier, consensus
// boolean, and error.
func ApplyFilter(filter []byte, reveals []RevealBody) ([]int, bool, error) {
	if len(filter) < 1 {
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}

	switch filter[0] {
	case filterNone:
		return make([]int, len(reveals)), true, nil

	case filterMode:
		dataPath, err := types.UnpackModeFilter(filter)
		if err != nil {
			return nil, false, err
		}

		if dataPath == "" {
			return nil, false, errors.New("empty JSON path in filter input [Mode]")
		}

		exitCodes := make([]uint8, 0, len(reveals))
		revealData := make([]string, 0, len(reveals))
		for _, r := range reveals {
			exitCodes = append(exitCodes, r.ExitCode)
			revealData = append(revealData, r.Reveal)
		}
		outliers, consensus := FilterMode(dataPath, exitCodes, revealData)
		return outliers, consensus, nil

	// TODO: Reactivate standard deviation filter
	// case filterStdDev:
	// 	return nil, false, errors.New("filter type standard deviation is not implemented")

	default:
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}
}

func FilterMode(dataPath string, exitCodes []uint8, reveals []string) ([]int, bool) {
	var nonZeroExitCode int
	vals := make([]string, 0, len(reveals))

	for i, r := range reveals {
		vals = append(vals, gjson.Get(r, dataPath).String())
		if exitCodes[i] != 0 {
			nonZeroExitCode++
		}
	}
	outliers, consensus := calcOutliers(vals, exitCodes)

	// If more than 1/3 reveals have non exit code, we reached a consensus that
	// these are unusable reveals.
	if consensus || nonZeroExitCode*3 > len(reveals) {
		return outliers, true
	}
	return outliers, false
}

func calcOutliers[T comparable](reveals []T, exitCode []uint8) ([]int, bool) {
	freq := make(map[T]int, len(reveals))
	outliers := make([]int, len(reveals))
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}

	var maxFreq int

	for i, r := range reveals {
		if exitCode[i] != 0 {
			continue
		}
		freq[r]++
		maxFreq = max(freq[r], maxFreq)
	}

	// If not MORE THAN 2/3 matches the max frequency, there is no data-consensus.
	if maxFreq*3 <= len(reveals)*2 {
		return outliers, false
	}

	for i, r := range reveals {
		if freq[r] != maxFreq || exitCode[i] != 0 {
			continue
		}
		outliers[i] = 0
	}

	return outliers, true
}
