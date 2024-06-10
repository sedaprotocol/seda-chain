package drfilters

import (
	"errors"

	"github.com/ethereum/go-ethereum/rlp"
)

const (
	None         = 0
	Mode         = 1
	StdDeviation = 2
)

type (
	tallyProp struct {
		Algo       uint
		JsonPath   string `rlp:"optional"`
		MaxSigma   uint64 `rlp:"optional"`
		NumberType uint8  `rlp:"optional"`
	}
)

// Outliers calculates which reveals are in acceptance criteria.
// It returns a list of True and False s. Where 0 means data at index i is an
// outlier.
func Outliers(tallyInput []byte, reveals map[string]interface{}) ([]bool, error) {
	var input tallyProp
	if err := rlp.DecodeBytes(tallyInput, &input); err != nil {
		return nil, err
	}

	outliers := make([]bool, 0, len(reveals))
	switch input.Algo {
	case None:
		for range reveals {
			outliers = append(outliers, true)
		}
	case Mode:
		return nil, errors.New("filter type Mode not implemented")
	case StdDeviation:
		return nil, errors.New("filter type Standard deviation not implemented")
	}
	return outliers, nil
}
