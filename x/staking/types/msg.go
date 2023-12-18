package types

import (
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

// // Validate validates the MsgCreateValidator sdk msg.
// func (msg MsgCreateValidatorWithVRF) Validate(ac address.Codec) error {
// 	// note that unmarshaling from bech32 ensures both non-empty and valid
// 	_, err := ac.StringToBytes(msg.ValidatorAddress)
// 	if err != nil {
// 		return sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
// 	}

// 	if msg.Pubkey == nil {
// 		return ErrEmptyValidatorPubKey
// 	}

// 	if !msg.Value.IsValid() || !msg.Value.Amount.IsPositive() {
// 		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid delegation amount")
// 	}

// 	if msg.Description == (Description{}) {
// 		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
// 	}

// 	if msg.Commission == (CommissionRates{}) {
// 		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty commission")
// 	}

// 	if err := msg.Commission.Validate(); err != nil {
// 		return err
// 	}

// 	if !msg.MinSelfDelegation.IsPositive() {
// 		return errorsmod.Wrap(
// 			sdkerrors.ErrInvalidRequest,
// 			"minimum self delegation must be a positive integer",
// 		)
// 	}

// 	if msg.Value.Amount.LT(msg.MinSelfDelegation) {
// 		return ErrSelfDelegationBelowMinimum
// 	}

// 	return nil
// }

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateValidatorWithVRF) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey, vrfPubKey cryptotypes.PubKey
	err := unpacker.UnpackAny(msg.Pubkey, &pubKey)
	if err != nil {
		return err
	}
	return unpacker.UnpackAny(msg.VrfPubkey, &vrfPubKey)
}
