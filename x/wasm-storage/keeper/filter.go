package keeper

import (
	"errors"

	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	filterTypeNone   byte = 0x00
	filterTypeMode   byte = 0x01
	filterTypeStdDev byte = 0x02
)

// ApplyFilter processes filter of the type specified in the first
// byte of consensus filter. It returns an outlier list, which is
// a boolean list where true at index i means that the reveal at
// index i is an outlier, consensus boolean, and error.
func ApplyFilter(input []byte, reveals []types.RevealBody) ([]int, bool, error) {
	// TODO: Return invalid filter input error instead?
	if len(input) < 1 {
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}

	var filter types.Filter
	var err error
	switch input[0] {
	case filterTypeNone:
		filter, err = types.NewFilterNone(input)
	case filterTypeMode:
		filter, err = types.NewFilterMode(input)
	case filterTypeStdDev:
		filter, err = types.NewFilterStdDev(input)
	default:
		// TODO: Return invalid filter input error instead?
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	outliers, err := filter.ApplyFilter(reveals)
	switch {
	case err == nil:
		return outliers, true, nil
	case errors.Is(err, types.ErrNoConsensus):
		return outliers, false, nil
	case errors.Is(err, types.ErrCorruptReveals):
		return allOutliers(len(reveals)), true, err
	default:
		return nil, false, err
	}
}

func allOutliers(length int) []int {
	outliers := make([]int, length)
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}
	return outliers
}
