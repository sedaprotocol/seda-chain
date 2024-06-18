package keeper

import (
	"errors"
)

const (
	filterNone   byte = 0x00
	filterMode   byte = 0x01
	filterStdDev byte = 0x02
)

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
		return nil, false, errors.New("filter type mode is not implemented")

	case filterStdDev:
		return nil, false, errors.New("filter type standard deviation is not implemented")

	default:
		return nil, false, errors.New("filter type is invalid")
	}
}
