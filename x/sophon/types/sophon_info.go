package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MaxMemoLength    = 3000
	DoNotModifyField = "[do-not-modify]"
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

	if s.Balance.IsNil() {
		return fmt.Errorf("balance is nil")
	}

	if s.UsedCredits.IsNil() {
		return fmt.Errorf("used credits is nil")
	}

	return nil
}
