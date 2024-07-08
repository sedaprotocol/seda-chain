package types

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func (m *MsgAddVRFKey) ValidateBasic(modules []string) error {
	if m.Pubkey == nil {
		return ErrNilValue.Wrap("msg.Pubkey")
	}
	if len(m.Name) < 3 {
		return ErrInvalidInput.Wrapf("len(%s) < 3", m.Name)
	}
	if !slices.Contains(modules, m.Application) {
		return ErrInvalidInput.Wrapf("module %s is not present in [%v]", m.Application, modules)
	}
	return nil
}

func (m *MsgAddVRFKey) PublicKey() (cryptotypes.PubKey, error) {
	pubKey, ok := m.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(ErrInvalidType, "%T is not a cryptotypes.PubKey", pubKey)
	}
	return pubKey, nil
}
