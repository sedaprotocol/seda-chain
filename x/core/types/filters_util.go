package types

import (
	"encoding/base64"
	"slices"

	"github.com/ohler55/ojg/gen"
	"github.com/ohler55/ojg/jp"
)

type dataAttributes struct {
	freqMap map[string]int // frequency map from data to frequency
	maxFreq int            // frequency of most frequent data in data list
}

// parseReveals parses a list of RevealBody objects using the given
// data path and returns a parsed data list along with its attributes.
// It also updates the given errors list to indicate true for the items
// that are corrupted. Note when an i-th reveal is corrupted, the i-th
// item in the data list is left as an empty string.
func parseReveals(reveals []Reveal, dataPath string, errors []bool) ([]string, dataAttributes) {
	var parser gen.Parser
	var maxFreq int
	freq := make(map[string]int, len(reveals))
	dataList := make([]string, len(reveals))
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
		obj, err := parser.Parse(revealBytes)
		if err != nil {
			errors[i] = true
			continue
		}
		expr, err := jp.ParseString(dataPath)
		if err != nil {
			errors[i] = true
			continue
		}
		elems := expr.GetNodes(obj)

		var data string
		switch len(elems) {
		case 0:
			errors[i] = true
			continue
		case 1:
			data = elems[0].String()
		default:
			elemsStr := make([]string, len(elems))
			for i, elem := range elems {
				elemsStr[i] = elem.String()
			}
			slices.Sort(elemsStr)
			data = elemsStr[0]
		}

		freq[data]++
		maxFreq = max(freq[data], maxFreq)
		dataList[i] = data
	}

	return dataList, dataAttributes{
		freqMap: freq,
		maxFreq: maxFreq,
	}
}
