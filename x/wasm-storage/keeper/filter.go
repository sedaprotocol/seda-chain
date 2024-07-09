package keeper

import (
	"encoding/base64"
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

		return filterMode(reveals, dataPath)

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

// filterMode takes in a list of reveals and a json path to extract
// the reveal value. It returns an outlier list, a consensus boolean,
// and an error if applicable.
// Value of 1 at index i in the outlier list indicates that the i-th
// reveal is an outlier. Value of 0 indicates a non-outlier reveal.
// A reveal is an outlier if it has less than or equal to 2/3 frequency.
// The consensus boolean is true if one of the following criteria is met:
// 1. More than 1/3 of the reveals are corrupted (non-zero exit code,
// invalid data path, etc.)
// 2. More than 2/3 of the reveals are identical.
func filterMode(reveals []RevealBody, dataPath string) ([]int, bool, error) {
	var maxFreq, corruptCount int
	freq := make(map[string]int, len(reveals))
	revealVals := make([]string, len(reveals))
	for i, r := range reveals {
		if r.ExitCode != 0 {
			corruptCount++
			continue
		}

		// Extract the reveal value to track frequency.
		revealBytes, err := base64.StdEncoding.DecodeString(r.Reveal)
		if err != nil {
			corruptCount++
			continue
		}
		res := gjson.GetBytes(revealBytes, dataPath)
		if !res.Exists() {
			corruptCount++
			continue
		}

		val := res.String()
		freq[val]++
		maxFreq = max(freq[val], maxFreq)
		revealVals[i] = val
	}

	outliers := make([]int, len(reveals))
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}

	// If more than 1/3 of the reveals are corrupted,
	// we reach consensus that the reveals are unusuable.
	if corruptCount*3 > len(reveals) {
		return outliers, true, nil
	}

	// If less than 2/3 of the reveals match the max frequency,
	// there is no consensus.
	if maxFreq*3 < len(reveals)*2 {
		return outliers, false, nil
	}

	for i, r := range revealVals {
		if freq[r] == maxFreq {
			outliers[i] = 0
		}
	}
	return outliers, true, nil
}
