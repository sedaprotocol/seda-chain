package keeper

import (
	"errors"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
	"github.com/tidwall/gjson"
)

const (
	filterAlgoNone byte = 0x00
	filterAlgoMode byte = 0x01
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
	case filterAlgoNone:
		return make([]int, len(reveals)), true, nil

	case filterAlgoMode:
		dataPath, err := types.UnpackModeFilter(filter)
		if err != nil {
			return nil, false, err
		}

		if dataPath == "" {
			return nil, false, errors.New("empty JSON path in filter input [Mode]")
		}

		outliers, consensus := filterMode(dataPath, reveals)
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

// filterMode takes in a list of reveals, and a json path to get the
// filtering data from the RevealBody->reveal.
// The function returns a list of int where value of each element can be
// either 0 or 1. At index i Value 1 means RevealBody[i] contains an
// outlier. If a reveal have MORE THAN 2/3 frequency, than it's not an
// outlier.
//
// The function also returns if a consensus is reached. Consensus can be
// reached in two ways.
// 1. More than 1/3 of the reveals are corrupted. i.e. has non-zero exit code.
// 2. More than 2/3 of the data is the same.
func filterMode(dataPath string, reveals []RevealBody) ([]int, bool) {
	var maxFreq, nonZeroExitCode int
	freq := make(map[string]int, len(reveals))
	revealVals := make([]string, 0, len(reveals))

	for _, r := range reveals {
		val := gjson.Get(r.Reveal, dataPath).String()
		revealVals = append(revealVals, val)
		if r.ExitCode != 0 {
			nonZeroExitCode++
			continue
		}
		freq[val]++
		maxFreq = max(freq[val], maxFreq)
	}

	outliers := make([]int, len(reveals))
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}

	// If more than 1/3 reveals have non exit code, we reached a consensus that
	// these are unusable reveals.
	if nonZeroExitCode*3 > len(reveals) {
		return outliers, true
	}

	// If less than 2/3 matches the max frequency, there is no data-consensus.
	if maxFreq*3 < len(reveals)*2 {
		return outliers, false
	}

	for i, r := range revealVals {
		if freq[r] != maxFreq || reveals[i].ExitCode != 0 {
			continue
		}
		outliers[i] = 0
	}
	return outliers, true
}
