package keeper

import (
	"encoding/base64"
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

// BuildFilter builds a filter based on the requestor-provided input.
func (k Keeper) BuildFilter(ctx sdk.Context, filterInput string, replicationFactor uint16) (types.Filter, error) {
	input, err := base64.StdEncoding.DecodeString(filterInput)
	if err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return nil, types.ErrInvalidFilterType
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	var filter types.Filter
	switch input[0] {
	case filterTypeNone:
		filter = types.NewFilterNone(params.FilterGasCostNone)
	case filterTypeMode:
		filter, err = types.NewFilterMode(input, params.FilterGasCostMultiplierMode, replicationFactor)
	case filterTypeStdDev:
		filter, err = types.NewFilterStdDev(input, params.FilterGasCostMultiplierStddev, replicationFactor)
	default:
		return nil, types.ErrInvalidFilterType
	}
	if err != nil {
		return nil, err
	}
	return filter, nil
}

// ApplyFilter processes filter of the type specified in the first
// byte of consensus filter. It returns an outlier list, which is
// a boolean list where true at index i means that the reveal at
// index i is an outlier, consensus boolean, consensus data proxy
// public keys, and error. It assumes that the reveals and their
// proxy public keys are sorted.
func ApplyFilter(filter types.Filter, reveals []types.RevealBody) (FilterResult, error) {
	var result FilterResult
	result.Outliers = make([]int, len(reveals))
	result.GasUsed = filter.GasCost()

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
