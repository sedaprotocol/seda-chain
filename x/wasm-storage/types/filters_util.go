package types

import (
	"encoding/base64"

	"github.com/tidwall/gjson"
)

// outlier list if ...
// CONTRACT: one of dataList and outliers is returned
// - DataList is returned only if ther is consensus
// dataAttributes?? dataAttrs

// ParseReveals parses the reveal bodies into a data list given a data path.
func ParseReveals(reveals []RevealBody, dataPath string) ([]string, []int, bool, map[string]int, int) {
	var maxFreq, corruptCount int
	freq := make(map[string]int, len(reveals))
	dataList := make([]string, len(reveals))
	for i, r := range reveals {
		if r.ExitCode != 0 {
			corruptCount++
			continue
		}

		// Extract the reveal data to track frequency.
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

		data := res.String()
		freq[data]++
		maxFreq = max(freq[data], maxFreq)
		dataList[i] = data
	}

	// If more than 1/3 of the reveals are corrupted,
	// we reach consensus that the reveals are unusable.
	if corruptCount*3 > len(reveals) {
		return nil, allOutlierList(len(reveals)), true, freq, maxFreq
	}

	// // If less than 2/3 of the reveals match the max frequency,
	// // there is no consensus.
	// if maxFreq*3 < len(reveals)*2 {
	// 	return nil, allOutlierList(len(reveals)), false, freq, maxFreq
	// }

	return dataList, nil, true, freq, maxFreq
}

func allOutlierList(length int) []int {
	outliers := make([]int, length)
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}
	return outliers
}
