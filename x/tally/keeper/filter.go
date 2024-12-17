package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sedaprotocol/seda-chain/x/tally/types"
)

const (
	filterTypeNone   byte = 0x00
	filterTypeMode   byte = 0x01
	filterTypeStdDev byte = 0x02
)

type FilterResult struct {
	Outliers     []int    // outlier list
	Consensus    bool     // whether consensus was reached
	ProxyPubKeys []string // consensus data proxy public keys
	GasUsed      uint64   // gas used for filter
}

// ApplyFilter processes filter of the type specified in the first
// byte of consensus filter. It returns an outlier list, which is
// a boolean list where true at index i means that the reveal at
// index i is an outlier, consensus boolean, consensus data proxy
// public keys, and error. It assumes that the reveals and their
// proxy public keys are sorted.
func (k Keeper) ApplyFilter(ctx sdk.Context, input []byte, reveals []types.RevealBody, replicationFactor int64) (FilterResult, error) {
	var result FilterResult
	result.Outliers = make([]int, len(reveals))

	if len(input) == 0 {
		return result, types.ErrInvalidFilterType
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
		return result, types.ErrNoBasicConsensus
	}
	result.ProxyPubKeys = proxyPubKeys

	params, err := k.GetParams(ctx)
	if err != nil {
		return result, err
	}

	var filter types.Filter
	switch input[0] {
	case filterTypeNone:
		filter, err = types.NewFilterNone(input)
		result.GasUsed = params.FilterGasCostNone
	case filterTypeMode:
		filter, err = types.NewFilterMode(input)
		result.GasUsed = params.FilterGasCostMultiplierMode * uint64(replicationFactor)
	case filterTypeStdDev:
		filter, err = types.NewFilterStdDev(input)
		result.GasUsed = params.FilterGasCostMultiplierStddev * uint64(replicationFactor)
	default:
		return result, types.ErrInvalidFilterType
	}
	if err != nil {
		result.GasUsed = 0
		return result, err
	}

	outliers, err := filter.ApplyFilter(reveals)
	switch {
	case err == nil:
		result.Outliers = outliers
		result.Consensus = true
		return result, nil
	case errors.Is(err, types.ErrNoConsensus):
		result.Outliers = outliers
		return result, nil
	case errors.Is(err, types.ErrCorruptReveals):
		return result, err
	default:
		return result, err
	}
}
