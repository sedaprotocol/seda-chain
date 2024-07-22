package types

import (
	"encoding/base64"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

type dataAttributes struct {
	freqMap map[any]int // frequency map from data to frequency
	maxFreq int         // frequency of most frequent data in data list
}

// parseReveals parses a list of RevealBody objects using a given data
// path and returns a data list. However, if more than 1/3 of the reveals
// are corrupted (i.e. cannot be parsed), no data list is returned and
// ErrCorruptReveals error is returned. When there is no error, it also
// returns dataAttributes struct since some filters require this information.
// Note that when an i-th reveal is corrupted, the i-th item in the data
// list is left as an empty string.
func parseReveals(reveals []RevealBody, dataPath string) ([]any, dataAttributes, error) {
	if len(reveals) == 0 {
		return nil, dataAttributes{}, ErrEmptyReveals
	}

	var maxFreq, corruptCount int
	freq := make(map[any]int, len(reveals))
	dataList := make([]any, len(reveals))
	for i, r := range reveals {
		if r.ExitCode != 0 {
			corruptCount++
			continue
		}

		revealBytes, err := base64.StdEncoding.DecodeString(r.Reveal)
		if err != nil {
			corruptCount++
			continue
		}
		obj, err := oj.Parse(revealBytes)
		if err != nil {
			corruptCount++
			continue
		}
		expr, err := jp.ParseString(dataPath)
		if err != nil {
			corruptCount++
			continue
		}
		elems := expr.Get(obj)
		if len(elems) < 1 {
			corruptCount++
			continue
		}
		data := elems[0]

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
