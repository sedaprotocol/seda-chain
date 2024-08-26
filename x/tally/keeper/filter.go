package keeper

import (
	"errors"
	"fmt"

	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

const (
	filterTypeNone   byte = 0x00
	filterTypeMode   byte = 0x01
	filterTypeStdDev byte = 0x02
)

// ApplyFilter processes filter of the type specified in the first
// byte of consensus filter. It returns an outlier list, which is
// a boolean list where true at index i means that the reveal at
// index i is an outlier, consensus boolean, and error. It assumes
// that the reveals and their proxy public keys are sorted.
func ApplyFilter(input []byte, reveals []types.RevealBody) ([]int, bool, error) {
	if len(input) == 0 {
		return make([]int, len(reveals)), false, types.ErrInvalidFilterType
	}

	// Determine consensus on tuple of (exit_code, proxy_pub_keys)
	var maxFreq int
	freq := make(map[string]int, len(reveals))
	for _, reveal := range reveals {
		var success bool
		if reveal.ExitCode != 0 {
			success = false
		}

		tuple := fmt.Sprintf("%v%v", success, reveal.ProxyPubKeys)
		freq[tuple]++
		maxFreq = max(freq[tuple], maxFreq)
	}
	if maxFreq*3 < len(reveals)*2 {
		return make([]int, len(reveals)), false, types.ErrNoBasicConsensus
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
