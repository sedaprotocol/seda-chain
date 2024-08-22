package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

const (
	MaxMemoLength    = 3000
	DoNotModifyField = "[do-not-modify]"
	UseMinimumDelay  = 0
)

func (p *ProxyConfig) Validate() error {
	if len(p.Memo) > MaxMemoLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid memo length; got: %d, max < %d", len(p.Memo), MaxMemoLength)
	}

	return nil
}

func (p *ProxyConfig) UpdateBasic(payoutAddress string, memo string) error {
	if payoutAddress != DoNotModifyField {
		p.PayoutAddress = payoutAddress
	}

	if memo != DoNotModifyField {
		p.Memo = memo
	}

	return p.Validate()
}
