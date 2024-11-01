package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddKey{}

func (m *MsgAddKey) Validate() error {
	if m.ValidatorAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty validator address")
	}
	return nil
}
