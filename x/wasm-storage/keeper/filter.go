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

const (
	INT uint8 = iota + 1
	INT8
	INT16
	INT32
	INT64
	UINT
	UINT8
	UINT16
	UINT32
	UINT64
	FLOAT32
	FLOAT64
)

type (
	filterProp struct {
		Algo       uint
		JSONPath   string `rlp:"optional"`
		MaxSigma   uint64 `rlp:"optional"`
		NumberType uint8  `rlp:"optional"`
	}
)

func FilterMode(jsonPath string, numberType uint8, exitCodes []uint8, reveals [][]byte) ([]bool, bool) {
	var (
		outliers        []bool
		consensus       bool
		nonZeroExitCode int
	)
	results := make([]gjson.Result, 0, len(reveals))

	for i, r := range reveals {
		results = append(results, gjson.GetBytes(r, jsonPath))
		if exitCodes[i] != 0 {
			nonZeroExitCode++
		}
	}

	switch numberType {
	case INT, INT8, INT16, INT32, INT64:
		vals := make([]int64, 0, len(reveals))
		for _, res := range results {
			vals = append(vals, res.Int())
		}
		outliers, consensus = calculate(vals)
	case UINT, UINT8, UINT16, UINT32, UINT64:
		vals := make([]uint64, 0, len(reveals))
		for _, res := range results {
			vals = append(vals, res.Uint())
		}
		outliers, consensus = calculate(vals)
	case FLOAT32, FLOAT64:
		vals := make([]float64, 0, len(reveals))
		for _, res := range results {
			vals = append(vals, res.Float())
		}
		outliers, consensus = calculate(vals)
	default:
		vals := make([]string, 0, len(reveals))
		for _, res := range results {
			vals = append(vals, res.String())
		}
		outliers, consensus = calculate(vals)
	}

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
		outliers, consensus = FilterMode(filter.JSONPath, filter.NumberType, exitCodes, revealData)
		return outliers, consensus, nil
	case StdDeviation:
		return nil, false, errors.New("filter type Standard deviation not implemented")
	}
	return outliers, true, nil
}
