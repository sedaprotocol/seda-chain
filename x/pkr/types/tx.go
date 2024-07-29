package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg                            = &MsgAddKey{}
	_ codectypes.UnpackInterfacesMessage = (*MsgAddKey)(nil)
)

func (m *MsgAddKey) Validate() error {
	// var pkAny *codectypes.Any
	// if m.Pubkey != nil {
	// 	var err error
	// 	if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
	// 		return nil, err
	// 	}
	// }

	if m.Pubkey == nil {
		return ErrNilValue.Wrap("msg.Pubkey")
	}
	return nil
}

func (m *MsgAddKey) PublicKey() (cryptotypes.PubKey, error) {
	pubKey, ok := m.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(ErrInvalidType, "%T is not a cryptotypes.PubKey", pubKey)
	}
	return pubKey, nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *MsgAddKey) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if m.Pubkey != nil {
		var pubKey cryptotypes.PubKey
		return unpacker.UnpackAny(m.Pubkey, &pubKey)
	}
	return nil
}
