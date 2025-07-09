package types

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appparams "github.com/sedaprotocol/seda-chain/app/params"
)

const (
	MaxUpdatesPerBlock       int    = 25
	DefaultMinFeeUpdateDelay uint32 = 73750 // Roughly 1 week with a ~8.2 sec block time
	LowestFeeUpdateDelay     uint32 = 1
)

var DefaultRegistrationFee = math.NewIntWithDecimal(50, 18) // (50)*10^(18) aseda

// DefaultParams returns default data-proxy module parameters.
func DefaultParams() Params {
	return Params{
		MinFeeUpdateDelay: DefaultMinFeeUpdateDelay,
		RegistrationFee:   sdk.NewCoin(appparams.DefaultBondDenom, DefaultRegistrationFee),
	}
}

// ValidateBasic performs basic validation on data-proxy module parameters.
func (p *Params) Validate() error {
	if p.MinFeeUpdateDelay < LowestFeeUpdateDelay {
		return sdkerrors.ErrInvalidRequest.Wrapf("MinFeeUpdateDelay lower than %d < %d", p.MinFeeUpdateDelay, LowestFeeUpdateDelay)
	}
	return nil
}
