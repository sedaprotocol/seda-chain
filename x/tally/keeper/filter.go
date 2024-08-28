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
// index i is an outlier, consensus boolean, consensus data proxy
// public keys, and error. It assumes that the reveals and their
// proxy public keys are sorted.
func ApplyFilter(input []byte, reveals []types.RevealBody) ([]int, bool, []string, error) {
	if len(input) == 0 {
		return make([]int, len(reveals)), false, []string{}, types.ErrInvalidFilterType
	}

	// Determine basic consensus on tuple of (exit_code, proxy_pub_keys)
	var maxFreq int
	var proxyPubKeys []string
	freq := make(map[string]int, len(reveals))
	for _, reveal := range reveals {
		success := reveal.ExitCode == 0
		tuple := fmt.Sprintf("%v%v", success, reveal.ProxyPubKeys)
		freq[tuple]++

		if freq[tuple] > maxFreq {
			proxyPubKeys = reveal.ProxyPubKeys
			maxFreq = freq[tuple]
		}
	}
	if maxFreq*3 < len(reveals)*2 {
		return make([]int, len(reveals)), false, proxyPubKeys, types.ErrNoBasicConsensus
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
		return make([]int, len(reveals)), false, proxyPubKeys, types.ErrInvalidFilterType
	}
	if err != nil {
		return make([]int, len(reveals)), false, proxyPubKeys, err
	}

	outliers, err := filter.ApplyFilter(reveals)
	switch {
	case err == nil:
		return outliers, true, proxyPubKeys, nil
	case errors.Is(err, types.ErrNoConsensus):
		return outliers, false, proxyPubKeys, nil
	case errors.Is(err, types.ErrCorruptReveals):
		return make([]int, len(reveals)), false, proxyPubKeys, err
	default:
		return make([]int, len(reveals)), false, proxyPubKeys, err
	}
}
