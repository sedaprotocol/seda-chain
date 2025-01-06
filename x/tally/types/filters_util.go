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

// parseReveals parses a list of RevealBody objects using the given
// data path and returns a parsed data list along with its attributes.
// It also updates the given errors list to indicate true for the items
// that are corrupted. Note when an i-th reveal is corrupted, the i-th
// item in the data list is left as nil.
func parseReveals(reveals []RevealBody, dataPath string, errors []bool) ([]any, dataAttributes) {
	var maxFreq int
	freq := make(map[any]int, len(reveals))
	dataList := make([]any, len(reveals))
	for i, r := range reveals {
		if r.ExitCode != 0 {
			errors[i] = true
			continue
		}

		revealBytes, err := base64.StdEncoding.DecodeString(r.Reveal)
		if err != nil {
			errors[i] = true
			continue
		}
		obj, err := oj.Parse(revealBytes)
		if err != nil {
			errors[i] = true
			continue
		}
		expr, err := jp.ParseString(dataPath)
		if err != nil {
			errors[i] = true
			continue
		}
		elems := expr.Get(obj)
		if len(elems) < 1 {
			errors[i] = true
			continue
		}
		data := elems[0]

		freq[data]++
		maxFreq = max(freq[data], maxFreq)
		dataList[i] = data
	}

	return dataList, dataAttributes{
		freqMap: freq,
		maxFreq: maxFreq,
	}
}
