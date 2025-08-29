package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MaxMemoLength      = 3000
	MaxPublicKeyLength = 66
	DoNotModifyField   = "[do-not-modify]"
)

func ValidateMemo(memo string) error {
	if len(memo) > MaxMemoLength {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid memo length; got: %d, max < %d", len(memo), MaxMemoLength)
	}

	return nil
}

func (s *SophonInfo) ValidateBasic() error {
	if err := ValidateMemo(s.Memo); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(s.OwnerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner address: %s", s.OwnerAddress)
	}

	if _, err := sdk.AccAddressFromBech32(s.AdminAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid admin address: %s", s.AdminAddress)
	}

	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", s.Address)
	}

	if len(s.PublicKey) == 0 {
		return fmt.Errorf("public key is empty")
	}

	if len(s.PublicKey) > MaxPublicKeyLength {
		return fmt.Errorf("public key is too long; got: %d, max < %d", len(s.PublicKey), MaxPublicKeyLength)
	}

	if s.Balance.IsNil() {
		return fmt.Errorf("balance is nil")
	}

	if s.Balance.IsNegative() {
		return fmt.Errorf("balance is negative")
	}

	if s.UsedCredits.IsNil() {
		return fmt.Errorf("used credits is nil")
	}

	if s.UsedCredits.IsNegative() {
		return fmt.Errorf("used credits is negative")
	}

	return nil
}
