package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg                            = &MsgAddKey{}
	_ codectypes.UnpackInterfacesMessage = (*MsgAddKey)(nil)
)

func (m *MsgAddKey) Validate() error {
	if m.ValidatorAddr == "" {
		return ErrEmptyValue.Wrap("empty validator address")
	}
	if m.PubKey == nil {
		return ErrEmptyValue.Wrap("empty public key")
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *MsgAddKey) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.PubKey != nil {
		var pubKey cryptotypes.PubKey
		return unpacker.UnpackAny(m.PubKey, &pubKey)
	}
	return nil
}
