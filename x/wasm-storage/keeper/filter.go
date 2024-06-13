package keeper

import (
	"errors"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tidwall/gjson"
)

const (
	None         = iota
	Mode         = 1
	StdDeviation = 2
)

type (
	filterProp struct {
		Algo       uint
		JSONPath   string `rlp:"optional"`
		MaxSigma   uint64 `rlp:"optional"`
		NumberType uint8  `rlp:"optional"`
	}
)

func FilterMode(jsonPath string, exitCodes []uint8, reveals [][]byte) ([]bool, bool) {
	var (
		outliers        []bool
		consensus       bool
		nonZeroExitCode int
	)

	vals := make([]string, 0, len(reveals))
	for i, r := range reveals {
		vals = append(vals, gjson.GetBytes(r, jsonPath).String())
		if exitCodes[i] != 0 {
			nonZeroExitCode++
		}
	}
	outliers, consensus = calculate(vals)

	if consensus || nonZeroExitCode*3 > 2*len(reveals) {
		return outliers, true
	}
	return outliers, false
}

func calculate[T comparable](reveals []T) ([]bool, bool) {
	freq := make(map[T]int, len(reveals))
	outliers := make([]bool, 0, len(reveals))
	var maxFreq int

	for _, r := range reveals {
		freq[r]++
		if freq[r] > maxFreq {
			maxFreq = freq[r]
		}
	}

	if maxFreq*3 < len(reveals)*2 {
		outliers = make([]bool, len(reveals))
		return outliers, false
	}

	for _, r := range reveals {
		outliers = append(outliers, freq[r] != maxFreq)
	}

	return outliers, true
}

// Outliers calculates which reveals are in acceptance criteria.
// It returns a list of True and False. Where True means data at index i is an
// outlier.
//
// Note: <param: tallyInput> is a rlp encoded and <param:reveals> is JSON serialized.
func Outliers(filterInput []byte, reveals []RevealBody) ([]bool, bool, error) {
	var filter filterProp
	if err := rlp.DecodeBytes(filterInput, &filter); err != nil {
		return nil, false, err
	}

	outliers := make([]bool, 0, len(reveals))
	var consensus bool
	switch filter.Algo {
	case None:
		for range reveals {
			outliers = append(outliers, false)
		}
	case Mode:
		if filter.JSONPath == "" {
			return nil, false, errors.New("empty JSON path")
		}

		exitCodes := make([]uint8, 0, len(reveals))
		revealData := make([][]byte, 0, len(reveals))
		for _, r := range reveals {
			exitCodes = append(exitCodes, r.ExitCode)
			revealData = append(revealData, r.Reveal)
		}
		outliers, consensus = FilterMode(filter.JSONPath, exitCodes, revealData)
		return outliers, consensus, nil
	case StdDeviation:
		return nil, false, errors.New("filter type Standard deviation not implemented")
	}
	return outliers, true, nil
}
