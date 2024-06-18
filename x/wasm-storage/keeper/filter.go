package keeper

import (
	"errors"

	"github.com/tidwall/gjson"
)

const (
	filterNone   byte = 0x00
	filterMode   byte = 0x01
	filterStdDev byte = 0x02
)

type (
	modeFilter struct {
		Algo     uint
		JSONPath string
	}

	//stdFilter struct {
	//	Algo       uint
	//	JSONPath   string
	//	MaxSigma   uint64
	//	NumberType uint8
	//}
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

// ApplyFilter processes filter of the type specified in the first byte of
// tally inputs. It returns an outlier list, which is a boolean list where
// true at index i means that the reveal at index i is an outlier, consensus
// boolean, and error.
func ApplyFilter(tallyInputs []byte, reveals []RevealBody) ([]bool, bool, error) {
	if len(tallyInputs) < 1 {
		return nil, false, errors.New("tally inputs should be at least 1 byte")
	}

	switch tallyInputs[0] {
	case filterNone:
		outliers := make([]bool, len(reveals))
		for i := range outliers {
			outliers[i] = false
		}
		return outliers, true, nil

	case filterMode:
		// TODO: Reactivate mode filter
		/*
			var filter modeFilter
			if err := rlp.DecodeBytes(tallyInputs, &filter); err != nil {
				return nil, false, err
			}
			if filter.JSONPath == "" {
				return nil, false, errors.New("empty JSON path")
			}

			exitCodes := make([]uint8, len(reveals))
			revealData := make([][]byte, len(reveals))
			for i, r := range reveals {
				exitCodes[i] = r.ExitCode
				revealData[i] = r.Reveal
			}
			outliers, consensus := FilterMode(filter.JSONPath, exitCodes, revealData)
			return outliers, consensus, nil
		*/
		return nil, false, errors.New("filter type mode is not implemented")

	case filterStdDev:
		return nil, false, errors.New("filter type standard deviation is not implemented")

	default:
		return nil, false, errors.New("filter type is invalid")
	}
}
