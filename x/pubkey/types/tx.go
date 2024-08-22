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
	for i, pair := range m.IndexedPubKeys {
		if pair.PubKey == nil {
			return ErrEmptyValue.Wrapf("empty public key at index %d", i)
		}
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *MsgAddKey) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, pair := range m.IndexedPubKeys {
		if pair.PubKey != nil {
			var pubKey cryptotypes.PubKey
			return unpacker.UnpackAny(pair.PubKey, &pubKey)
		}
	}
	return nil
}
