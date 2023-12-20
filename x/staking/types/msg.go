package types

import (
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewMsgCreateValidator creates a new MsgCreateValidator instance.
// Delegator address and validator address are the same.
func NewMsgCreateValidatorWithVRF(
	valAddr string, pubKey cryptotypes.PubKey, vrfPubKey cryptotypes.PubKey,
	selfDelegation sdk.Coin, description types.Description, commission types.CommissionRates, minSelfDelegation math.Int,
) (*MsgCreateValidatorWithVRF, error) {
	var pkAny *codectypes.Any
	if pubKey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubKey); err != nil {
			return nil, err
		}
	}

	var vrfPkAny *codectypes.Any
	if vrfPubKey != nil {
		var err error
		if vrfPkAny, err = codectypes.NewAnyWithValue(vrfPubKey); err != nil {
			return nil, err
		}
	}

	return &MsgCreateValidatorWithVRF{
		Description:       description,
		ValidatorAddress:  valAddr,
		Pubkey:            pkAny,
		VrfPubkey:         vrfPkAny,
		Value:             selfDelegation,
		Commission:        commission,
		MinSelfDelegation: minSelfDelegation,
	}, nil
}

// Validate validates the MsgCreateValidatorWithVRF sdk msg.
func (msg MsgCreateValidatorWithVRF) Validate(ac address.Codec) error {
	if msg.Pubkey == nil || msg.VrfPubkey == nil {
		return fmt.Errorf("empty validator public key or VRF public key")
	}

	msgCreateVal := &types.MsgCreateValidator{
		Description:       msg.Description,
		Commission:        msg.Commission,
		MinSelfDelegation: msg.MinSelfDelegation,
		ValidatorAddress:  msg.ValidatorAddress,
		Pubkey:            msg.Pubkey,
		Value:             msg.Value,
	}
	return msgCreateVal.Validate(ac)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidatorWithVRF) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey, vrfPubKey cryptotypes.PubKey
	err := unpacker.UnpackAny(msg.Pubkey, &pubKey)
	if err != nil {
		return err
	}
	return unpacker.UnpackAny(msg.VrfPubkey, &vrfPubKey)
}
