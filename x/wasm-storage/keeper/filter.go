package keeper

import (
	"github.com/sedaprotocol/seda-chain/x/wasm-storage/types"
)

const (
	filterTypeNone   byte = 0x00
	filterTypeMode   byte = 0x01
	filterTypeStdDev byte = 0x02
)

// ApplyFilter processes filter of the type specified in the first byte of
// consensus filter. It returns an outlier list, which is a boolean list where
// true at index i means that the reveal at index i is an outlier, consensus
// boolean, and error.
func ApplyFilter(input []byte, reveals []types.RevealBody) ([]int, bool, error) {
	// TODO: Return error?
	if len(input) < 1 {
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}

	var filter types.Filter
	switch input[0] {
	case filterTypeNone:
		return make([]int, len(reveals)), true, nil

	case filterTypeMode:
		filter = new(types.FilterMode)

	case filterTypeStdDev:
		filter = new(types.FilterStdDev)

	// TODO: Return error?
	default:
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}

	err := filter.DecodeFilterInput(input)
	if err != nil {
		return nil, false, err
	}
	return filter.ApplyFilter(reveals)
}
