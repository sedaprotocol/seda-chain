package types

import (
	"sort"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg                            = &MsgAddKey{}
	_ codectypes.UnpackInterfacesMessage = (*MsgAddKey)(nil)
)

func (m *MsgAddKey) Validate() error {
	if m.ValidatorAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("empty validator address")
	}

	sort.Slice(m.IndexedPubKeys, func(i, j int) bool {
		return m.IndexedPubKeys[i].Index < m.IndexedPubKeys[j].Index
	})
	for i, pair := range m.IndexedPubKeys {
		if i > 0 {
			if pair.Index == m.IndexedPubKeys[i-1].Index {
				return sdkerrors.ErrInvalidRequest.Wrapf("duplicate index at %d", pair.Index)
			}
		}
		if pair.PubKey == nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("empty public key at index %d", pair.Index)
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
