package types

import (
	"encoding/base64"

	"github.com/tidwall/gjson"
)

type dataAttributes struct {
	freqMap map[string]int // frequency map from data to frequency
	maxFreq int            // frequency of most frequent data in data list
}

// parseReveals parses a list of RevealBody objects using a given
// data path and returns a data list. However, if more than 1/3 of
// the reveals are corrupted (i.e. cannot be parsed), no data list
// is returned and ErrCorruptReveals error is returned. When there
// is no error, it also returns dataAttributes struct since some
// filters require this information.
func parseReveals(reveals []RevealBody, dataPath string) ([]string, dataAttributes, error) {
	var maxFreq, corruptCount int
	freq := make(map[string]int, len(reveals))
	dataList := make([]string, len(reveals))
	for i, r := range reveals {
		if r.ExitCode != 0 {
			corruptCount++
			continue
		}

		revealBz, err := base64.StdEncoding.DecodeString(r.Reveal)
		if err != nil {
			corruptCount++
			continue
		}
		res := gjson.GetBytes(revealBz, dataPath)
		if !res.Exists() {
			corruptCount++
			continue
		}

		data := res.String()
		freq[data]++
		maxFreq = max(freq[data], maxFreq)
		dataList[i] = data
	}

	if corruptCount*3 > len(reveals) {
		return nil, dataAttributes{}, ErrCorruptReveals
	}
	return dataList, dataAttributes{
		freqMap: freq,
		maxFreq: maxFreq,
	}, nil
}
