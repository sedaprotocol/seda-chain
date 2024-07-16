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
	if len(input) < 1 {
		return make([]int, len(reveals)), false, types.ErrInvalidFilterType
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
		return make([]int, len(reveals)), false, types.ErrInvalidFilterType
	}
	if err != nil {
		return make([]int, len(reveals)), false, err
	}

	outliers, err := filter.ApplyFilter(reveals)
	switch {
	case err == nil:
		return outliers, true, nil
	case errors.Is(err, types.ErrNoConsensus):
		return outliers, false, nil
	case errors.Is(err, types.ErrCorruptReveals):
		return make([]int, len(reveals)), false, err
	default:
		return make([]int, len(reveals)), false, err
	}
}
